package controllers

const (
	ConfigPrefix      = "harbor-controller"
	ReconciliationKey = ConfigPrefix + "-max-reconcile"
	WatchChildrenKey  = ConfigPrefix + "-watch-children"
	HarborClassKey    = ConfigPrefix + "-class"
)

const (
	DefaultConcurrentReconcile = 1
	DefaultWatchChildren       = true
	DefaultHarborClass         = ""
)
