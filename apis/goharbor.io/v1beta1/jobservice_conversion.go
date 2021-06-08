package v1beta1

import (
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *JobService) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.JobService)

	dst.ObjectMeta = src.ObjectMeta
	dst.Status = src.Status

	Convert_v1beta1_JobServiceSpec_To_v1alpha3_JobServiceSpec(&src.Spec, &dst.Spec)

	return nil
}

func (dst *JobService) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.JobService)

	dst.ObjectMeta = src.ObjectMeta
	dst.Status = src.Status

	Convert_v1alpha3_JobServiceSpec_To_v1beta1_JobServiceSpec(&src.Spec, &dst.Spec)

	return nil
}

func Convert_v1beta1_JobServiceSpec_To_v1alpha3_JobServiceSpec(src *JobServiceSpec, dst *v1alpha3.JobServiceSpec) {
	dst.ComponentSpec = src.ComponentSpec
	dst.SecretRef = src.SecretRef
	dst.TLS = src.TLS
	dst.Proxy = src.Proxy

	Convert_v1beta1_JobServiceTokenSpec_To_v1alpha3_JobServiceTokenSpec(&src.TokenService, &dst.TokenService)

	Convert_v1beta1_JobServiceCoreSpec_To_v1alpha3_JobServiceCoreSpec(&src.Core, &dst.Core)

	Convert_v1beta1_JobServicePoolSpec_To_v1alpha3_JobServicePoolSpec(&src.WorkerPool, &dst.WorkerPool)

	Convert_v1beta1_JobServiceLoggerConfigSpec_To_v1alpha3_JobServiceLoggerConfigSpec(&src.JobLoggers, &dst.JobLoggers)

	Convert_v1beta1_JobServiceLoggerConfigSpec_To_v1alpha3_JobServiceLoggerConfigSpec(&src.Loggers, &dst.Loggers)

	Convert_v1beta1_RegistryControllerConnectionSpec_To_v1alpha3_RegistryControllerConnectionSpec(&src.Registry, &dst.Registry)

}

func Convert_v1beta1_JobServiceTokenSpec_To_v1alpha3_JobServiceTokenSpec(src *JobServiceTokenSpec, dst *v1alpha3.JobServiceTokenSpec) {
	dst.URL = src.URL
}

func Convert_v1beta1_JobServiceCoreSpec_To_v1alpha3_JobServiceCoreSpec(src *JobServiceCoreSpec, dst *v1alpha3.JobServiceCoreSpec) {
	dst.SecretRef = src.SecretRef
	dst.URL = src.URL

}

func Convert_v1beta1_JobServicePoolSpec_To_v1alpha3_JobServicePoolSpec(src *JobServicePoolSpec, dst *v1alpha3.JobServicePoolSpec) {
	dst.WorkerCount = src.WorkerCount
	dst.Backend = src.Backend
	dst.Redis = v1alpha3.JobServicePoolRedisSpec{
		RedisConnection: src.Redis.RedisConnection,
		Namespace:       src.Redis.Namespace,
		IdleTimeout:     src.Redis.IdleTimeout,
	}
}

func Convert_v1beta1_JobServiceLoggerConfigSpec_To_v1alpha3_JobServiceLoggerConfigSpec(src *JobServiceLoggerConfigSpec, dst *v1alpha3.JobServiceLoggerConfigSpec) {
	if src.Database != nil {
		dst.Database = &v1alpha3.JobServiceLoggerConfigDatabaseSpec{
			Level:   src.Database.Level,
			Sweeper: src.Database.Sweeper,
		}
	}

	if src.STDOUT != nil {
		dst.STDOUT = &v1alpha3.JobServiceLoggerConfigSTDOUTSpec{
			Level: src.STDOUT.Level,
		}
	}

	if len(src.Files) != 0 {
		dst.Files = make([]v1alpha3.JobServiceLoggerConfigFileSpec, 0)
		for _, file := range src.Files {
			dst.Files = append(dst.Files, v1alpha3.JobServiceLoggerConfigFileSpec{
				Volume:  file.Volume,
				Level:   file.Level,
				Sweeper: file.Sweeper,
			})
		}
	}
}

func Convert_v1beta1_RegistryControllerConnectionSpec_To_v1alpha3_RegistryControllerConnectionSpec(src *RegistryControllerConnectionSpec, dst *v1alpha3.RegistryControllerConnectionSpec) {
	dst.RegistryURL = src.RegistryURL
	dst.ControllerURL = src.ControllerURL
	dst.Credentials = v1alpha3.CoreComponentsRegistryCredentialsSpec{
		Username:    src.Credentials.Username,
		PasswordRef: src.Credentials.PasswordRef,
	}
}

func Convert_v1alpha3_JobServiceSpec_To_v1beta1_JobServiceSpec(src *v1alpha3.JobServiceSpec, dst *JobServiceSpec) {
	dst.ComponentSpec = src.ComponentSpec
	dst.SecretRef = src.SecretRef
	dst.TLS = src.TLS
	dst.Proxy = src.Proxy

	Convert_v1alpha3_JobServiceTokenSpec_To_v1beta1_JobServiceTokenSpec(&src.TokenService, &dst.TokenService)

	Convert_v1alpha3_JobServiceCoreSpec_To_v1beta1_JobServiceCoreSpec(&src.Core, &dst.Core)

	Convert_v1alpha3_JobServicePoolSpec_To_v1beta1_JobServicePoolSpec(&src.WorkerPool, &dst.WorkerPool)

	Convert_v1alpha3_JobServiceLoggerConfigSpec_To_v1beta1_JobServiceLoggerConfigSpec(&src.JobLoggers, &dst.JobLoggers)

	Convert_v1alpha3_JobServiceLoggerConfigSpec_To_v1beta1_JobServiceLoggerConfigSpec(&src.Loggers, &dst.Loggers)

	Convert_v1alpha3_RegistryControllerConnectionSpec_To_v1beta1_RegistryControllerConnectionSpec(&src.Registry, &dst.Registry)

}

func Convert_v1alpha3_JobServiceTokenSpec_To_v1beta1_JobServiceTokenSpec(src *v1alpha3.JobServiceTokenSpec, dst *JobServiceTokenSpec) {
	dst.URL = src.URL
}

func Convert_v1alpha3_JobServiceCoreSpec_To_v1beta1_JobServiceCoreSpec(src *v1alpha3.JobServiceCoreSpec, dst *JobServiceCoreSpec) {
	dst.SecretRef = src.SecretRef
	dst.URL = src.URL

}

func Convert_v1alpha3_JobServicePoolSpec_To_v1beta1_JobServicePoolSpec(src *v1alpha3.JobServicePoolSpec, dst *JobServicePoolSpec) {
	dst.WorkerCount = src.WorkerCount
	dst.Backend = src.Backend
	dst.Redis = JobServicePoolRedisSpec{
		RedisConnection: src.Redis.RedisConnection,
		Namespace:       src.Redis.Namespace,
		IdleTimeout:     src.Redis.IdleTimeout,
	}
}

func Convert_v1alpha3_JobServiceLoggerConfigSpec_To_v1beta1_JobServiceLoggerConfigSpec(src *v1alpha3.JobServiceLoggerConfigSpec, dst *JobServiceLoggerConfigSpec) {
	if src.Database != nil {
		dst.Database = &JobServiceLoggerConfigDatabaseSpec{
			Level:   src.Database.Level,
			Sweeper: src.Database.Sweeper,
		}
	}

	if src.STDOUT != nil {
		dst.STDOUT = &JobServiceLoggerConfigSTDOUTSpec{
			Level: src.STDOUT.Level,
		}
	}

	if len(src.Files) != 0 {
		dst.Files = make([]JobServiceLoggerConfigFileSpec, 0)
		for _, file := range src.Files {
			dst.Files = append(dst.Files, JobServiceLoggerConfigFileSpec{
				Volume:  file.Volume,
				Level:   file.Level,
				Sweeper: file.Sweeper,
			})
		}
	}
}

func Convert_v1alpha3_RegistryControllerConnectionSpec_To_v1beta1_RegistryControllerConnectionSpec(src *v1alpha3.RegistryControllerConnectionSpec, dst *RegistryControllerConnectionSpec) {
	dst.RegistryURL = src.RegistryURL
	dst.ControllerURL = src.ControllerURL
	dst.Credentials = CoreComponentsRegistryCredentialsSpec{
		Username:    src.Credentials.Username,
		PasswordRef: src.Credentials.PasswordRef,
	}
}
