package harborcluster

import (
	"github.com/go-logr/logr"
	"github.com/goharbor/harbor-operator/controllers/goharbor/harborcluster/cache"
	"github.com/goharbor/harbor-operator/controllers/goharbor/harborcluster/database"
	"github.com/goharbor/harbor-operator/controllers/goharbor/harborcluster/harbor"
	"github.com/goharbor/harbor-operator/controllers/goharbor/harborcluster/storage"
	"github.com/goharbor/harbor-operator/pkg/image"
	"github.com/goharbor/harbor-operator/pkg/k8s"
	"github.com/goharbor/harbor-operator/pkg/lcm"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
)

type ServiceGetter interface {
	// For Redis
	Cache() lcm.Controller

	// For database
	Database() lcm.Controller

	// For storage
	Storage() lcm.Controller

	// For harbor itself
	Harbor() lcm.Controller
}

type GetOptions struct {
	Client      k8s.Client
	Recorder    record.EventRecorder
	Log         logr.Logger
	DClient     k8s.DClient
	Scheme      *runtime.Scheme
	ImageGetter image.Getter
}

type ServiceGetterImpl struct {
	redisReconciler    *cache.RedisReconciler
	databaseReconciler *database.PostgreSQLReconciler
	storageReconciler  *storage.MinIOReconciler
	harborReconciler   *harbor.HarborReconciler
}

func NewServiceGetterImpl(options *GetOptions) ServiceGetter {
	// TODO need update
	return &ServiceGetterImpl{
		redisReconciler:    &cache.RedisReconciler{},
		databaseReconciler: &database.PostgreSQLReconciler{},
		storageReconciler:  &storage.MinIOReconciler{},
		harborReconciler:   &harbor.HarborReconciler{},
	}
}

func (impl *ServiceGetterImpl) Harbor() lcm.Controller {
	return impl.harborReconciler
}

func (impl *ServiceGetterImpl) Cache() lcm.Controller {
	return impl.redisReconciler
}

func (impl *ServiceGetterImpl) Database() lcm.Controller {
	return impl.databaseReconciler
}

func (impl *ServiceGetterImpl) Storage() lcm.Controller {
	return impl.storageReconciler
}
