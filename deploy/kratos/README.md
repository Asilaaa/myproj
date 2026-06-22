# Ory Kratos social login: Google + GitHub

This project's frontend already renders Ory self-service flow nodes, including OIDC provider buttons. To enable Google and GitHub login, configure Kratos with OIDC providers and mapper files.

## 1. Create OAuth apps

### Google
Create an OAuth client in Google Cloud Console.

Use this redirect URI:

```text
https://sukoon.uz/ory/self-service/methods/oidc/callback/google
```

If you still test with the raw IP, temporarily use:

```text
http://34.66.123.47/ory/self-service/methods/oidc/callback/google
```

Recommended scopes:
- `openid`
- `email`
- `profile`

### GitHub
Create an OAuth App in GitHub settings.

Use this callback URL:

```text
https://sukoon.uz/ory/self-service/methods/oidc/callback/github
```

Recommended scopes:
- `user:email`

## 2. Copy mapper files to the server

```bash
sudo mkdir -p /etc/kratos
sudo cp /opt/myproj/deploy/kratos/google.jsonnet /etc/kratos/google.jsonnet
sudo cp /opt/myproj/deploy/kratos/github.jsonnet /etc/kratos/github.jsonnet
sudo chown root:kratos /etc/kratos/google.jsonnet /etc/kratos/github.jsonnet
sudo chmod 640 /etc/kratos/google.jsonnet /etc/kratos/github.jsonnet
```

## 3. Add OIDC providers to `/etc/kratos/kratos.yml`

You can start from `deploy/kratos/kratos.yml.example` in this repository and then fill in secrets on the server.

Under `selfservice.methods`, keep password enabled and add `oidc` like this:

```yaml
selfservice:
  methods:
    password:
      enabled: true
    oidc:
      enabled: true
      config:
        providers:
          - id: google
            provider: google
            client_id: GOOGLE_CLIENT_ID
            client_secret: GOOGLE_CLIENT_SECRET
            mapper_url: file:///etc/kratos/google.jsonnet
            scope:
              - openid
              - email
              - profile

          - id: github
            provider: github
            client_id: GITHUB_CLIENT_ID
            client_secret: GITHUB_CLIENT_SECRET
            mapper_url: file:///etc/kratos/github.jsonnet
            scope:
              - user:email
```

Replace the client IDs and client secrets with your real values.

## 4. Restart Kratos

```bash
sudo systemctl restart kratos
sudo systemctl status kratos --no-pager -l
```

## 5. Test

Open:

```text
https://sukoon.uz/login
```

or

```text
https://sukoon.uz/registration
```

You should now see extra provider buttons for Google and GitHub.

## Notes

- The frontend does not need special per-provider code; Ory flow nodes are rendered generically.
- Production should use HTTPS callback URLs such as `https://sukoon.uz/...`.
- If a provider returns no usable email, login/registration may fail because the identity schema requires `traits.email`.
