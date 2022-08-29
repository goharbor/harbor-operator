package consts

const (
	// AnnotationHarborServer is the annnotation for harbor server.
	AnnotationHarborServer = "goharbor.io/harbor"
	// AnnotationAccount is the annnotation for service account.
	AnnotationAccount = "goharbor.io/service-account"
	// AnnotationProject is the annnotation for harbor project name.
	AnnotationProject = "goharbor.io/project"
	// AnnotationRobot is the annotation for robot id.
	AnnotationRobot = "goharbor.io/robot"
	// AnnotationRobotSecretRef is the annotation for robot secret reference.
	AnnotationRobotSecretRef = "goharbor.io/robot-secret" //nolint:gosec
	// AnnotationSecOwner is the annotation for owner.
	AnnotationSecOwner = "goharbor.io/owner"
	// AnnotationImageRewriteRuleConfigMapRef is the annotation for reference to configmap that stores rules.
	AnnotationImageRewriteRuleConfigMapRef = "goharbor.io/rewriting-rules"

	// ConfigMapKeyHarborServer is the key in configmap that for HSC.
	ConfigMapKeyHarborServer = "hsc"
	// ConfigMapKeyRules is the key in configmap that for rules.
	ConfigMapKeyRules = "rules"
	// ConfigMapKeyRewriting is the key in configmap that for whether turn on image rewrite.
	ConfigMapKeyRewriting = "rewriting"
	// ConfigMapValueRewritingOff is the key in configmap that for rewrite to turn off.
	ConfigMapValueRewritingOff = "off"
	// ConfigMapValueRewritingOn is the key in configmap that for rewrite to turn on.
	ConfigMapValueRewritingOn = "on"
)
