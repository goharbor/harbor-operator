# Customize Images

There are several ways provided for you to customize the images pulled by the Harbor operator.

## By operator environment variables

Set the image pulling path of each Harbor component or dependent service via the operator environment variables when deploying the operator controller into the Kubernetes cluster. The image paths set by this way will be applicable to all the Harbor clusters deployed from the operator.

By default, the environment variables settings of the image pulling paths are not activated in the default [full stack](../manifests/cluster/deployment.yaml) operator deployment manifest and [Harbor operator only](../manifests/harbor/deployment.yaml) deployment manifest. You can follow the way shown below to enable it.

1. Uncomment the following content section in the [full stack](../manifests/cluster/kustomization.yaml) or [Harbor operator only](../manifests/harbor/kustomization.yaml) kustomization yaml file.

  ```yaml
  # - patch/image-source.yaml
  ```

1. Regenerate the deployment manifest with the updated kustomization file and apply it

   ```shell
   # Change workdir
   cd manifests/cluster
  
   # Generate the deployment manifest
   kustomize build . -o my_new_deployment.yaml
  
   # Apply
   kubectl apply -f my_new_deployment.yaml
  
   # Or use command pipe
   kustomize build . | kubectl apply -f -
   ```

or directly apply via kustomization file

  ```shell
  # Change workdir
  cd manifests/cluster
  
  # Apply with kustomization template file
  kubectl apply -k kustomization.yaml
  ```

## Image source of HarborCluster CR

Configure `spec.imageSource` to specify the general image source from where pulling images. The image settings configured here are applicable to all the components of the deploying Harbor.

For how to configure it, refer to the [CRD spec](./CRD/custom-resource-definition.md#configure-image-source)

## Component images of HarborCluster CR

Each component including both Harbor components and its in-cluster dependent services has related configurations to specify image pulling source for the specified component.

For how to configure the image source for the specified Harbor component, refer to the portal section of the [CRD spec](./CRD/custom-resource-definition.md#harbor-component-related-fields).

For how to configure the image source for the specified in-cluster dependent service, refer to the in-cluster storage section of [CRD spec](./CRD/custom-resource-definition.md#in-cluster-storage-configuration-inclusterstorage).
