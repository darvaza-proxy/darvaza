# Helpers for implementing http.Handlers

## Content Negotiation

### QualityList

The QualityList parser allows choosing the best option during Content Negotiation, e.g. accepted `Content-Type`s.

### BestQuality

`qlist` offers two helpers to choose the best option from a QualityList and a list of
supported options, `BestQuality()` and `BestQualityWithIdentity()`. _Identity_ is an special
option we consider unless it's explicitly forbidden.

### BestEncoding

`qlist.BestEncoding()` is a special case of `BestQualityWithIdentity()` using the `Accept`
header, and falling back to `"identity"` as magic type.

### See also

* [Accept](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Accept)
* [Content Negotiation](https://developer.mozilla.org/en-US/docs/Web/HTTP/Content_negotiation)
* [Quality Values](https://developer.mozilla.org/en-US/docs/Glossary/Quality_values)
