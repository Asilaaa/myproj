local claims = std.extVar('claims');

{
  identity: {
    traits: {
      email: claims.email,
      name: {
        first: if std.objectHas(claims, 'given_name') && claims.given_name != null then claims.given_name else '',
        last: if std.objectHas(claims, 'family_name') && claims.family_name != null then claims.family_name else '',
      },
    },
  },
}
