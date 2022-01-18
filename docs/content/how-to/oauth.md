# OAuth configuration

Guest cluster kube admin password will be exposed only when user has explicitly not specified the OAuth configuration.
```
      apiVersion: config.openshift.io/v1
      kind: OAuth
      metadata:
        name: cluster
      spec:
        identityProviders:
        - github:
            clientID: 123456789
            clientSecret:
              name: github-identity
            organizations:
            - example-org
          mappingMethod: claim
          name: github
          type: GitHub
    secretRefs:
    - name: fancydomain-servingcert
    - name: github-identity
```