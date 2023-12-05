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

Para instalar o operador com o IBM Common Services Operator, siga as instruções de instalação e configuração no IBM Knowledge Center.

- Se você estiver usando o operador como parte de um IBM Cloud Pak, consulte a documentação desse IBM Cloud Pak. Para obter uma lista de IBM Cloud Paks, consulte [IBM Cloud Paks que usam serviços comuns](http://ibm.biz/cpcs_cloudpaks).
- Se você estiver usando o operador com um IBM Containerized Software, consulte o IBM Cloud Platform Common Services Knowledge Center [documentação do instalador](http://ibm.biz/cpcs_opinstall).

## SecurityContextConstraints Requirements

O serviço IBM Common Services MongoDB suporta a execução com as Restrições de Contexto de Segurança (SCCs) restritas padrão do OpenShift Container Platform 4.3.

Definição personalizada de SecurityContextConstraints:

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

Para obter mais informações sobre as restrições de contexto de segurança do OpenShift Container Platform, consulte [Gerenciando restrições de contexto de segurança](https://docs.openshift.com/container-platform/4.3/authentication/managing-security-context-constraints.html).

### Características principais

**_Admin Secret_**

A partir da versão 1.1.0, agora você pode fornecer seu próprio segredo `icp-mongodb-admin`. O segredo deve ter um campo `user` e um campo `password` e estar no mesmo namespace onde o mongoDB será criado. Se você optar por não fornecer um segredo, um usuário e uma senha aleatórios serão criados e usados. O segredo `icp-mongodb-admin` persistirá após a desinstalação ou remoção do recurso personalizado MongoDB para que a desinstalação e a reinstalação sejam possíveis usando os mesmos volumes persistentes.

Exemplo de YAML para criar seu próprio segredo de administrador antes da instalação. O usuário e a senha são criptografados em base64.
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

### Notas

Ele foi projetado para uso somente pelo IBM Common Services.

A operadora não oferece suporte à atualização do CR na versão 1.0.0. Para fazer alterações em uma instância implantada do MongoDB, é melhor editar diretamente o statefulset.

Ao implantar o MongoDB, é melhor usar 3 réplicas, especialmente se você não estiver fazendo backup de seus dados. É possível que os dados sejam corrompidos e a recuperação de uma implantação de 3 réplicas seja muito mais fácil.
