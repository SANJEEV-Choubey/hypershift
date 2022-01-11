---
title: OAuth configuration
---

# OAuth configuration
Kude admin password will be exposed only when user has explicitly not specified the OAuth configuration.

Create oauth client and secret

```
oc create -f <(echo '
kind: OAuthClient
apiVersion: oauth.openshift.io/v1
metadata:
 name: demo
secret: "password"" 
redirectURIs:
 - "http://www.example.com/"
grantMethod: auto
')
```

