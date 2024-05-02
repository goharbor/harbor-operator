package project

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"

	"github.com/go-logr/logr"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/models"
	"github.com/pkg/errors"
	goharborv1 "github.com/plotly/harbor-operator/apis/goharbor.io/v1beta1"
)

type memberUpdate struct {
	desired *models.ProjectMember
	current *models.ProjectMemberEntity
}

type memberDifferences struct {
	update []memberUpdate
	create []*models.ProjectMember
	delete []*models.ProjectMemberEntity
}

const (
	harborAPIProjectAdminRole int = 1
	harborAPIDeveloperRole    int = 2
	harborAPIGuestRole        int = 3
	harborAPIMaintainerRole   int = 4
)

// map string role mappings from CRD to int for Harbor API.
var memberRoleMapping = map[string]int{
	"projectAdmin": harborAPIProjectAdminRole,
	"developer":    harborAPIDeveloperRole,
	"guest":        harborAPIGuestRole,
	"maintainer":   harborAPIMaintainerRole,
}

func (r *Reconciler) reconcileMembership(hp *goharborv1.HarborProject, log logr.Logger) (err error) { //nolint:funlen
	// get current project members from Harbor API
	currentMemberships, err := r.Harbor.GetProjectMembers(hp)
	if err != nil {
		return err
	}

	// detect changes via hash from status field to skip unnecessary list comparisons
	previousHash := hp.Status.MembershipHash

	currentHash, err := generateHash(currentMemberships, hp.Spec.HarborProjectMemberships)
	if err != nil {
		return err
	}

	if previousHash == currentHash {
		// no changes, finish reconcile
		return nil
	}

	log.Info("reconcile membership, changes detected.", "previousHash", previousHash, "currentHash", currentHash)

	// create Harbor API objects for desired memberships defined in custom resource
	desiredMemberships, err := createDesiredMemberships(hp.Spec.HarborProjectMemberships)
	if err != nil {
		return err
	}

	// check length of current/desired member arrays, end reconcile if both are empty.
	currentMembershipsCnt := len(currentMemberships)
	desiredMembershipsCnt := len(desiredMemberships)

	if currentMembershipsCnt == 0 && desiredMembershipsCnt == 0 {
		log.Info("Nothing to do.", "current members", currentMembershipsCnt, "desired members", desiredMembershipsCnt)

		return nil
	}

	log.Info("Start reconcile", "current members", currentMembershipsCnt, "desired members", desiredMembershipsCnt)

	// find differences between current and desired members.
	differences := findDifferences(currentMemberships, desiredMemberships, log)

	err = r.updateMemberships(hp, differences, log)
	if err != nil {
		return err
	}

	// update hash a final time
	currentMemberships, err = r.Harbor.GetProjectMembers(hp)
	if err != nil {
		return err
	}

	hp.Status.MembershipHash, err = generateHash(currentMemberships, hp.Spec.HarborProjectMemberships)
	if err != nil {
		return err
	}

	log.Info("Membership reconcile complete.", "project", hp.Spec.ProjectName)

	return nil
}

func findDifferences(currentMemberships []*models.ProjectMemberEntity, desiredMemberships []models.ProjectMember, log logr.Logger) *memberDifferences {
	differences := memberDifferences{
		update: []memberUpdate{},
		create: []*models.ProjectMember{},
		delete: []*models.ProjectMemberEntity{},
	}

	desiredMembershipsCnt := len(desiredMemberships)
	currentMembershipsCnt := len(currentMemberships)

	// first, sort member slices for binary search
	sort.Slice(currentMemberships, func(i, j int) bool {
		return currentMemberships[i].EntityName < currentMemberships[j].EntityName
	})
	sort.Slice(desiredMemberships, func(i, j int) bool {
		return getProjectMemberName(&desiredMemberships[i]) < getProjectMemberName(&desiredMemberships[j])
	})

	// search all currentMembers in desiredMembers. If found, mark for update or deletion if necessary.
	for _, currentMember := range currentMemberships {
		idx := sort.Search(desiredMembershipsCnt, func(i int) bool {
			return getProjectMemberName(&desiredMemberships[i]) >= currentMember.EntityName
		})
		if idx < desiredMembershipsCnt && areMembersEqual(currentMember, &desiredMemberships[idx]) && currentMember.RoleID != desiredMemberships[idx].RoleID {
			log.Info("found matching members with differences, mark for update", "member", currentMember.EntityName)

			differences.update = append(differences.update, memberUpdate{desired: &desiredMemberships[idx], current: currentMember})
		} else if idx == desiredMembershipsCnt || getProjectMemberName(&desiredMemberships[idx]) != currentMember.EntityName {
			log.Info("currentMember was not found in desiredMemberships, mark for deletion.", "member", currentMember.EntityName)

			differences.delete = append(differences.delete, currentMember)
		}
	}

	// search all desiredMembers in currentMembers. If not found, mark for creation.
	for i := range desiredMemberships {
		desiredMemberName := getProjectMemberName(&desiredMemberships[i])

		idx := sort.Search(currentMembershipsCnt, func(i int) bool {
			return currentMemberships[i].EntityName >= desiredMemberName
		})

		if idx == currentMembershipsCnt || currentMemberships[idx].EntityName != desiredMemberName {
			log.Info("desiredMember was not found in currentMemberships, mark for creation.", "member", desiredMemberName)

			differences.create = append(differences.create, &desiredMemberships[i])
		}
	}

	log.Info("finished planning project member reconcile.", "create", len(differences.create), "update", len(differences.update), "delete", len(differences.delete))

	return &differences
}

func (r *Reconciler) updateMemberships(p *goharborv1.HarborProject, differences *memberDifferences, log logr.Logger) error {
	// delete all members marked for deletion
	for _, delMember := range differences.delete {
		log.Info("delete project member", "member", delMember.EntityName)

		err := r.Harbor.DeleteProjectMember(p.Spec.ProjectName, delMember.ID)
		if err != nil {
			return err
		}
	}

	// create all members marked for creation
	for _, createMember := range differences.create {
		name := getProjectMemberName(createMember)

		log.Info("create project member", "member", name)

		err := r.Harbor.CreateProjectMember(p.Spec.ProjectName, createMember)
		if err != nil {
			return err
		}
	}

	// update all members marked for updating
	for _, updateMember := range differences.update {
		log.Info("update project member", "member", updateMember.current.EntityName)

		err := r.Harbor.UpdateProjectMember(p.Spec.ProjectName, updateMember.current.ID, &models.RoleRequest{RoleID: updateMember.desired.RoleID})
		if err != nil {
			return err
		}
	}

	return nil
}

func areMembersEqual(harborMember *models.ProjectMemberEntity, k8sMember *models.ProjectMember) bool {
	return harborMember.EntityType == "g" && k8sMember.MemberGroup != nil && k8sMember.MemberGroup.GroupName == harborMember.EntityName ||
		harborMember.EntityType == "u" && k8sMember.MemberUser != nil && k8sMember.MemberUser.Username == harborMember.EntityName
}

func getProjectMemberName(member *models.ProjectMember) (name string) {
	switch {
	case member.MemberGroup != nil:
		return member.MemberGroup.GroupName
	case member.MemberUser != nil:
		return member.MemberUser.Username
	default:
		return ""
	}
}

func createDesiredMemberships(definedMemberships []*goharborv1.HarborProjectMember) ([]models.ProjectMember, error) {
	desiredMembers := []models.ProjectMember{}

	for _, definedMember := range definedMemberships {
		newMember := models.ProjectMember{}

		switch definedMember.Type {
		case "group":
			newMember.MemberGroup = &models.UserGroup{GroupName: definedMember.Name}
		case "user":
			newMember.MemberUser = &models.UserEntity{Username: definedMember.Name}
		default:
			return nil, errors.Errorf("unexpected member type \"%s\" for member \"%s\"", definedMember.Type, definedMember.Name)
		}

		newMember.RoleID = int64(memberRoleMapping[definedMember.Role])
		desiredMembers = append(desiredMembers, newMember)
	}

	return desiredMembers, nil
}

// marshal all current and desired memberships into json and hash them.
// this hash is used to efficiently find differences later on.
func generateHash(currentMemberships []*models.ProjectMemberEntity, desiredMemberships []*goharborv1.HarborProjectMember) (string, error) {
	type membershipComp struct {
		CurrentMemberships []*models.ProjectMemberEntity
		DesiredMemberships []*goharborv1.HarborProjectMember
	}

	membershipByteArr, err := json.Marshal(membershipComp{CurrentMemberships: currentMemberships, DesiredMemberships: desiredMemberships})
	if err != nil {
		err = errors.Wrap(err, "error marshaling memberships for comparison")

		return "", err
	}

	currentHashArr := sha256.Sum256(membershipByteArr)

	return hex.EncodeToString(currentHashArr[:]), nil
}
