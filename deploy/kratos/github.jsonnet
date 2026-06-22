local claims = std.extVar('claims');

{
  identity: {
    traits: {
      email: claims.email,
      name: {
        first: if std.objectHas(claims, 'login') && claims.login != null then claims.login else '',
        last: '',
      },
    },
  },
}
