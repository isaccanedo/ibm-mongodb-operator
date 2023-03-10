# ibm-mongodb-operator

O operador `ibm-mongodb-operator` é construído para suportar o IBM Cloud Platform Common Services. Ele suporta um banco de dados mongoDB que é compartilhado pelos serviços dentro do IBM Cloud Platform Common Services.

## Plataformas suportadas

Red Hat OpenShift Container Platform 4.2 ou mais recente instalado em uma das seguintes plataformas:

   - Linux x86_64
   - Linux on Power (ppc64le)
   - Linux on IBM Z and LinuxONE

## Versões do operador

| Versão | Data | Detatalhes |
| ----- | ---- | ----------------- |
| 1.1.0 | Julho 2020 | Permitir que os usuários configurem seu próprio segredo administrativo </br> - O CSV define as dependências que ele deve executar
| 1.0.0 | Março 2020 | Oferta inicial do operador MongoDB

## Pré-requisitos

Antes de instalar este operador, você precisa primeiro instalar as dependências e pré-requisitos do operador:

- Para obter a lista de dependências do operador, consulte o IBM Knowledge Center [documentação de dependências de Serviços Comuns](http://ibm.biz/cpcs_opdependencies).

- Para obter a lista de pré-requisitos para instalar o operador, consulte o IBM Knowledge Center [Preparando para instalar a documentação de serviços](http://ibm.biz/cpcs_opinstprereq).

## Documentatação

To install the operator with the IBM Common Services Operator follow the installation and configuration instructions within the IBM Knowledge Center.

- If you are using the operator as part of an IBM Cloud Pak, see the documentation for that IBM Cloud Pak. For a list of IBM Cloud Paks, see [IBM Cloud Paks that use Common Services](http://ibm.biz/cpcs_cloudpaks).
- If you are using the operator with an IBM Containerized Software, see the IBM Cloud Platform Common Services Knowledge Center [Installer documentation](http://ibm.biz/cpcs_opinstall).

## SecurityContextConstraints Requirements

The IBM Common Services MongoDB service supports running with the OpenShift Container Platform 4.3 default restricted Security Context Constraints (SCCs).

Custom SecurityContextConstraints definition:

```
allowHostDirVolumePlugin: false
allowHostIPC: false
allowHostNetwork: false
allowHostPID: false
allowHostPorts: false
allowPrivilegeEscalation: true
allowPrivilegedContainer: false
allowedCapabilities: null
apiVersion: security.openshift.io/v1
defaultAddCapabilities: null
fsGroup:
  type: MustRunAs
groups:
- system:authenticated
kind: SecurityContextConstraints
metadata:
  annotations:
    kubernetes.io/description: restricted denies access to all host features and requires
      pods to be run with a UID, and SELinux context that are allocated to the namespace.  This
      is the most restrictive SCC and it is used by default for authenticated users.
  creationTimestamp: "2020-06-17T15:06:39Z"
  generation: 1
  name: restricted
  resourceVersion: "6161"
  selfLink: /apis/security.openshift.io/v1/securitycontextconstraints/restricted
  uid: 255a542b-b0ac-11ea-97cc-00000a104120
priority: null
readOnlyRootFilesystem: false
requiredDropCapabilities:
- KILL
- MKNOD
- SETUID
- SETGID
runAsUser:
  type: MustRunAsRange
seLinuxContext:
  type: MustRunAs
supplementalGroups:
  type: RunAsAny
users: []
volumes:
- configMap
- downwardAPI
- emptyDir
- persistentVolumeClaim
- projected
- secret
```

For more information about the OpenShift Container Platform Security Context Constraints, see [Managing Security Context Constraints](https://docs.openshift.com/container-platform/4.3/authentication/managing-security-context-constraints.html).

### Key Features

**_Admin Secret_**

Starting with version 1.1.0, you can now supply your own `icp-mongodb-admin` secret. The secret must have a `user` field and a `password` field and be in the same namespace where mongoDB is going to be created. If you choose not to supply a secret, a random user and password will be created and used. The `icp-mongodb-admin` secret will persist after uninstalling or removing the MongoDB custom resource so that uninstall and re-install are possible by using the same Persistent Volumes.

Example YAML for creating your own admin secret before installation. The user and password are base64 encrypted.
```
apiVersion: v1
kind: Secret
metadata:
  name: icp-mongodb-admin
  namespace: ibm-common-services
type: Opaque
data:
  password: SFV6a2NYMkdKa2tBZA==
  user: dGpOcDR5Unc=
```

### Notes

This is designed for use by IBM Common Services only.

The operator does not support updating the CR in version 1.0.0. To make changes to a deployed MongoDB instance, it is best to directly edit the statefulset.

When you deploy MongoDB, it is better to use 3 replicas, especially if you are not backing up your data. It is possible for the data to be corrupted and recovering from a 3-replica deployment is much easier.
