## Documentation [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#section-documentation "Go to Documentation")

### Overview [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#pkg-overview "Go to Overview")

Package protocol contains data structures and validation functionality outlined in the Web Authentication specification ([https://www.w3.org/TR/webauthn](https://www.w3.org/TR/webauthn)). The data structures here attempt to conform as much as possible to their definitions, but some structs (like those that are used as part of validation steps) contain additional fields that help us unpack and validate the data we unmarshall. When implementing this library, most developers will primarily be using the API outlined in the webauthn package.

### Index [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#pkg-index "Go to Index")

- [Constants](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#pkg-constants)
- [Variables](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#pkg-variables)
- [func FullyQualifiedOrigin(rawOrigin string) (fqOrigin string, err error)](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#FullyQualifiedOrigin)
- [func IsOriginInHaystack(needle string, haystack []string) bool](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#IsOriginInHaystack)
- [func RegisterAttestationFormat(format AttestationFormat, handler attestationFormatValidationHandler)](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#RegisterAttestationFormat)
- [func ResidentKeyNotRequired() \*bool](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ResidentKeyNotRequired)
- [func ResidentKeyRequired() \*bool](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ResidentKeyRequired)
- [type AppleAnonymousAttestation](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AppleAnonymousAttestation)
- [type AttestationFormat](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AttestationFormat)
- [type AttestationObject](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AttestationObject)
- - [func (a \*AttestationObject) Verify(relyingPartyID string, clientDataHash []byte, userVerificationRequired bool, ...) (err error)](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AttestationObject.Verify)
  - [func (a \*AttestationObject) VerifyAttestation(clientDataHash []byte, mds metadata.Provider) (err error)](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AttestationObject.VerifyAttestation)
- [type AttestedCredentialData](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AttestedCredentialData)
- [type AuthenticationExtensions](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticationExtensions)
- [type AuthenticationExtensionsClientOutputs](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticationExtensionsClientOutputs)
- [type AuthenticatorAssertionResponse](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorAssertionResponse)
- [type AuthenticatorAttachment](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorAttachment)
- [type AuthenticatorAttestationResponse](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorAttestationResponse)
- - [func (ccr *AuthenticatorAttestationResponse) Parse() (p *ParsedAttestationResponse, err error)](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorAttestationResponse.Parse)
- [type AuthenticatorData](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorData)
- - [func (a \*AuthenticatorData) Unmarshal(rawAuthData []byte) (err error)](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorData.Unmarshal)
  - [func (a \*AuthenticatorData) Verify(rpIdHash []byte, appIDHash []byte, userVerificationRequired bool, ...) (err error)](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorData.Verify)
- [type AuthenticatorFlags](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorFlags)
- - [func (flag AuthenticatorFlags) HasAttestedCredentialData() bool](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorFlags.HasAttestedCredentialData)
  - [func (flag AuthenticatorFlags) HasBackupEligible() bool](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorFlags.HasBackupEligible)
  - [func (flag AuthenticatorFlags) HasBackupState() bool](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorFlags.HasBackupState)
  - [func (flag AuthenticatorFlags) HasExtensions() bool](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorFlags.HasExtensions)
  - [func (flag AuthenticatorFlags) HasUserPresent() bool](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorFlags.HasUserPresent)
  - [func (flag AuthenticatorFlags) HasUserVerified() bool](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorFlags.HasUserVerified)
  - [func (flag AuthenticatorFlags) UserPresent() bool](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorFlags.UserPresent)
  - [func (flag AuthenticatorFlags) UserVerified() bool](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorFlags.UserVerified)
- [type AuthenticatorResponse](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorResponse)
- [type AuthenticatorSelection](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorSelection)
- [type AuthenticatorTransport](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorTransport)
- [type CeremonyType](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#CeremonyType)
- [type CollectedClientData](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#CollectedClientData)
- - [func (c \*CollectedClientData) Verify(storedChallenge string, ceremony CeremonyType, ...) (err error)](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#CollectedClientData.Verify)
- [type ConveyancePreference](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ConveyancePreference)
- [type Credential](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#Credential)
- [type CredentialAssertion](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#CredentialAssertion)
- [type CredentialAssertionResponse](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#CredentialAssertionResponse)
- - [func (car CredentialAssertionResponse) Parse() (par \*ParsedCredentialAssertionData, err error)](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#CredentialAssertionResponse.Parse)
- [type CredentialCreation](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#CredentialCreation)
- [type CredentialCreationResponse](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#CredentialCreationResponse)
- - [func (ccr CredentialCreationResponse) Parse() (pcc \*ParsedCredentialCreationData, err error)](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#CredentialCreationResponse.Parse)
- [type CredentialDescriptor](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#CredentialDescriptor)
- [type CredentialEntity](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#CredentialEntity)
- [type CredentialMediationRequirement](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#CredentialMediationRequirement)
- [type CredentialParameter](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#CredentialParameter)
- [type CredentialType](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#CredentialType)
- [type Error](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#Error)
- - [func ValidateMetadata(ctx context.Context, mds metadata.Provider, aaguid uuid.UUID, ...) (protoErr \*Error)](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ValidateMetadata)
- - [func (e \*Error) Error() string](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#Error.Error)
  - [func (e \*Error) Unwrap() error](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#Error.Unwrap)
  - [func (e *Error) WithDetails(details string) *Error](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#Error.WithDetails)
  - [func (e *Error) WithError(err error) *Error](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#Error.WithError)
  - [func (e *Error) WithInfo(info string) *Error](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#Error.WithInfo)
- [type Extensions](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#Extensions)
- [type ParsedAssertionResponse](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ParsedAssertionResponse)
- [type ParsedAttestationResponse](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ParsedAttestationResponse)
- [type ParsedCredential](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ParsedCredential)
- [type ParsedCredentialAssertionData](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ParsedCredentialAssertionData)
- - [func ParseCredentialRequestResponse(response *http.Request) (*ParsedCredentialAssertionData, error)](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ParseCredentialRequestResponse)
  - [func ParseCredentialRequestResponseBody(body io.Reader) (par \*ParsedCredentialAssertionData, err error)](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ParseCredentialRequestResponseBody)
  - [func ParseCredentialRequestResponseBytes(data []byte) (par \*ParsedCredentialAssertionData, err error)](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ParseCredentialRequestResponseBytes)
- - [func (p \*ParsedCredentialAssertionData) Verify(storedChallenge string, relyingPartyID string, ...) error](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ParsedCredentialAssertionData.Verify)
- [type ParsedCredentialCreationData](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ParsedCredentialCreationData)
- - [func ParseCredentialCreationResponse(request *http.Request) (*ParsedCredentialCreationData, error)](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ParseCredentialCreationResponse)
  - [func ParseCredentialCreationResponseBody(body io.Reader) (pcc \*ParsedCredentialCreationData, err error)](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ParseCredentialCreationResponseBody)
  - [func ParseCredentialCreationResponseBytes(data []byte) (pcc \*ParsedCredentialCreationData, err error)](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ParseCredentialCreationResponseBytes)
- - [func (pcc \*ParsedCredentialCreationData) Verify(storedChallenge string, verifyUser bool, verifyUserPresence bool, ...) (clientDataHash []byte, err error)](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ParsedCredentialCreationData.Verify)
- [type ParsedPublicKeyCredential](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ParsedPublicKeyCredential)
- - [func (ppkc ParsedPublicKeyCredential) GetAppID(authExt AuthenticationExtensions, credentialAttestationType string) (appID string, err error)](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ParsedPublicKeyCredential.GetAppID)
- [type PublicKeyCredential](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#PublicKeyCredential)
- [type PublicKeyCredentialCreationOptions](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#PublicKeyCredentialCreationOptions)
- [type PublicKeyCredentialHints](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#PublicKeyCredentialHints)
- [type PublicKeyCredentialRequestOptions](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#PublicKeyCredentialRequestOptions)
- - [func (a \*PublicKeyCredentialRequestOptions) GetAllowedCredentialIDs() [][]byte](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#PublicKeyCredentialRequestOptions.GetAllowedCredentialIDs)
- [type RelyingPartyEntity](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#RelyingPartyEntity)
- [type ResidentKeyRequirement](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ResidentKeyRequirement)
- [type SafetyNetResponse](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#SafetyNetResponse)
- [type ServerResponse](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ServerResponse)
- [type ServerResponseStatus](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ServerResponseStatus)
- [type TokenBinding](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#TokenBinding)
- [type TokenBindingStatus](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#TokenBindingStatus)
- [type TopOriginVerificationMode](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#TopOriginVerificationMode)
- [type URLEncodedBase64](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#URLEncodedBase64)
- - [func CreateChallenge() (challenge URLEncodedBase64, err error)](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#CreateChallenge)
- - [func (e URLEncodedBase64) MarshalJSON() ([]byte, error)](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#URLEncodedBase64.MarshalJSON)
  - [func (e URLEncodedBase64) String() string](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#URLEncodedBase64.String)
  - [func (e \*URLEncodedBase64) UnmarshalJSON(data []byte) error](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#URLEncodedBase64.UnmarshalJSON)
- [type UserEntity](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#UserEntity)
- [type UserVerificationRequirement](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#UserVerificationRequirement)

### Constants [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#pkg-constants "Go to Constants")

[View Source](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/attestation_androidkey.go#L210)

```
const (
	Verified verifiedBootState = iota
	SelfSigned
	Unverified
	Failed
)
```

[View Source](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/attestation_androidkey.go#L217)

```
const (
	// KM_ORIGIN_GENERATED means generated in keymaster. Should not exist outside the TEE.
	KM_ORIGIN_GENERATED = iota

	// KM_ORIGIN_DERIVED means derived inside keymaster. Likely exists off-device.
	KM_ORIGIN_DERIVED

	// KM_ORIGIN_IMPORTED means imported into keymaster. Existed as clear text in Android.
	KM_ORIGIN_IMPORTED

	// KM_ORIGIN_UNKNOWN means keymaster did not record origin.  This value can only be seen on keys in a keymaster0
	// implementation. The keymaster0 adapter uses this value to document the fact that it is unknown whether the key
	// was generated inside or imported into keymaster.
	KM_ORIGIN_UNKNOWN
)
```

[View Source](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/attestation_androidkey.go#L234)

```
const (
	// KM_PURPOSE_ENCRYPT is usable with RSA, EC and AES keys.
	KM_PURPOSE_ENCRYPT = iota

	// KM_PURPOSE_DECRYPT is usable with RSA, EC and AES keys.
	KM_PURPOSE_DECRYPT

	// KM_PURPOSE_SIGN is usable with RSA, EC and HMAC keys.
	KM_PURPOSE_SIGN

	// KM_PURPOSE_VERIFY is usable with RSA, EC and HMAC keys.
	KM_PURPOSE_VERIFY

	// KM_PURPOSE_DERIVE_KEY is usable with EC keys.
	KM_PURPOSE_DERIVE_KEY

	// KM_PURPOSE_WRAP is usable with wrapped keys.
	KM_PURPOSE_WRAP
)
```

[View Source](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/extensions.go#L10)

```
const (
	ExtensionAppID        = "appid"
	ExtensionAppIDExclude = "appidExclude"
)
```

[View Source](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/challenge.go#L8)

```
const ChallengeLength = 32
```

ChallengeLength - Length of bytes to generate for a challenge.

[View Source](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/credential.go#L263)

```
const (
	CredentialTypeFIDOU2F = "fido-u2f"
)
```

### Variables [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#pkg-variables "Go to Variables")

[View Source](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/errors.go#L46)

```
var (
	ErrBadRequest = &Error{
		Type:    "invalid_request",
		Details: "Error reading the request data",
	}
	ErrChallengeMismatch = &Error{
		Type:    "challenge_mismatch",
		Details: "Stored challenge and received challenge do not match",
	}
	ErrParsingData = &Error{
		Type:    "parse_error",
		Details: "Error parsing the authenticator response",
	}
	ErrAuthData = &Error{
		Type:    "auth_data",
		Details: "Error verifying the authenticator data",
	}
	ErrVerification = &Error{
		Type:    "verification_error",
		Details: "Error validating the authenticator response",
	}
	ErrAttestation = &Error{
		Type:    "attestation_error",
		Details: "Error validating the attestation data provided",
	}
	ErrInvalidAttestation = &Error{
		Type:    "invalid_attestation",
		Details: "Invalid attestation data",
	}
	ErrMetadata = &Error{
		Type:    "invalid_metadata",
		Details: "",
	}
	ErrAttestationFormat = &Error{
		Type:    "invalid_attestation",
		Details: "Invalid attestation format",
	}
	ErrAttestationCertificate = &Error{
		Type:    "invalid_certificate",
		Details: "Invalid attestation certificate",
	}
	ErrAssertionSignature = &Error{
		Type:    "invalid_signature",
		Details: "Assertion Signature against auth data and client hash is not valid",
	}
	ErrUnsupportedKey = &Error{
		Type:    "invalid_key_type",
		Details: "Unsupported Public Key Type",
	}
	ErrUnsupportedAlgorithm = &Error{
		Type:    "unsupported_key_algorithm",
		Details: "Unsupported public key algorithm",
	}
	ErrNotSpecImplemented = &Error{
		Type:    "spec_unimplemented",
		Details: "This field is not yet supported by the WebAuthn spec",
	}
	ErrNotImplemented = &Error{
		Type:    "not_implemented",
		Details: "This field is not yet supported by this library",
	}
)
```

### Functions [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#pkg-functions "Go to Functions")

#### func [FullyQualifiedOrigin](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/client.go#L60) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#FullyQualifiedOrigin "Go to FullyQualifiedOrigin")

```
func FullyQualifiedOrigin(rawOrigin string) (fqOrigin string, err error)
```

FullyQualifiedOrigin returns the origin per the HTML spec: (scheme)://(host)[:(port)].

#### func [IsOriginInHaystack](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/client.go#L220) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#IsOriginInHaystack "Go to IsOriginInHaystack")added in **v0.14.0**

```
func IsOriginInHaystack(needle string, haystack []string) bool
```

IsOriginInHaystack checks if the needle is in the haystack using the mechanism to determine origin equality defined in HTML5 Section 5.3 and RFC3986 Section 6.2.1.

Specifically if the needle value has the 'http://' or 'https://' prefix (case-insensitive) and can be parsed as a URL; we check each item in the haystack to see if it matches the same rules, and then if the scheme and host (with a normalized port) components match case-insensitively then they're considered a match.

If the needle value does not have the 'http://' or 'https://' prefix (case-insensitive) or can't be parsed as a URL equality is determined using simple string comparison.

It is important to note that this function completely ignores Apple Associated Domains entirely as Apple is using an unassigned Well-Known URI in breech of Well-Known Uniform Resource Identifiers (RFC8615).

See (Origin Definition): [https://www.w3.org/TR/2011/WD-html5-20110525/origin-0.html](https://www.w3.org/TR/2011/WD-html5-20110525/origin-0.html)

See (Simple String Comparison Definition): [https://datatracker.ietf.org/doc/html/rfc3986#section-6.2.1](https://datatracker.ietf.org/doc/html/rfc3986#section-6.2.1)

See (Apple Associated Domains): [https://developer.apple.com/documentation/xcode/supporting-associated-domains](https://developer.apple.com/documentation/xcode/supporting-associated-domains)

See (IANA Well Known URI Assignments): [https://www.iana.org/assignments/well-known-uris/well-known-uris.xhtml](https://www.iana.org/assignments/well-known-uris/well-known-uris.xhtml)

See (Well-Known Uniform Resource Identifiers): [https://datatracker.ietf.org/doc/html/rfc8615](https://datatracker.ietf.org/doc/html/rfc8615)

#### func [RegisterAttestationFormat](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/attestation.go#L87) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#RegisterAttestationFormat "Go to RegisterAttestationFormat")

```
func RegisterAttestationFormat(format AttestationFormat, handler attestationFormatValidationHandler)
```

RegisterAttestationFormat is a method to register attestation formats with the library. Generally using one of the locally registered attestation formats is enough.

#### func [ResidentKeyNotRequired](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/authenticator.go#L385) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ResidentKeyNotRequired "Go to ResidentKeyNotRequired")added in **v0.2.0**

```
func ResidentKeyNotRequired() *bool
```

ResidentKeyNotRequired - Do not require that the private key be resident to the client device.

#### func [ResidentKeyRequired](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/authenticator.go#L378) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ResidentKeyRequired "Go to ResidentKeyRequired")

```
func ResidentKeyRequired() *bool
```

ResidentKeyRequired - Require that the key be private key resident to the client device.

### Types [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#pkg-types "Go to Types")

#### type [AppleAnonymousAttestation](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/attestation_apple.go#L88) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AppleAnonymousAttestation "Go to AppleAnonymousAttestation")

```
type AppleAnonymousAttestation struct {
	Nonce []byte `asn1:"tag:1,explicit"`
}
```

AppleAnonymousAttestation represents the attestation format for Apple, who have not yet published a schema for the extension (as of JULY 2021.)

#### type [AttestationFormat](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/options.go#L194) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AttestationFormat "Go to AttestationFormat")added in **v0.11.0**

```
type AttestationFormat string
```

AttestationFormat is an internal representation of the relevant inputs for registration.

Specification: §5.4 Options for Credential Creation ([https://w3c.github.io/webauthn/#dom-publickeycredentialcreationoptions-attestationformats](https://w3c.github.io/webauthn/#dom-publickeycredentialcreationoptions-attestationformats)) Registry: [https://www.iana.org/assignments/webauthn/webauthn.xhtml](https://www.iana.org/assignments/webauthn/webauthn.xhtml)

```
const (
	// AttestationFormatPacked is the "packed" attestation statement format is a WebAuthn-optimized format for
	// attestation. It uses a very compact but still extensible encoding method. This format is implementable by
	// authenticators with limited resources (e.g., secure elements).
	AttestationFormatPacked AttestationFormat = "packed"

	// AttestationFormatTPM is the TPM attestation statement format returns an attestation statement in the same format
	// as the packed attestation statement format, although the rawData and signature fields are computed differently.
	AttestationFormatTPM AttestationFormat = "tpm"

	// AttestationFormatAndroidKey is the attestation statement format for platform authenticators on versions "N", and
	// later, which may provide this proprietary "hardware attestation" statement.
	AttestationFormatAndroidKey AttestationFormat = "android-key"

	// AttestationFormatAndroidSafetyNet is the attestation statement format that Android-based platform authenticators
	// MAY produce an attestation statement based on the Android SafetyNet API.
	AttestationFormatAndroidSafetyNet AttestationFormat = "android-safetynet"

	// AttestationFormatFIDOUniversalSecondFactor is the attestation statement format that is used with FIDO U2F
	// authenticators.
	AttestationFormatFIDOUniversalSecondFactor AttestationFormat = "fido-u2f"

	// AttestationFormatApple is the attestation statement format that is used with Apple devices' platform
	// authenticators.
	AttestationFormatApple AttestationFormat = "apple"

	// AttestationFormatNone is the attestation statement format that is used to replace any authenticator-provided
	// attestation statement when a WebAuthn Relying Party indicates it does not wish to receive attestation information.
	AttestationFormatNone AttestationFormat = "none"
)
```

#### type [AttestationObject](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/attestation.go#L67) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AttestationObject "Go to AttestationObject")

```
type AttestationObject struct {
	// The authenticator data, including the newly created public key. See [AuthenticatorData] for more info.
	AuthData AuthenticatorData

	// The byteform version of the authenticator data, used in part for signature validation.
	RawAuthData []byte `json:"authData"`

	// The format of the Attestation data.
	Format string `json:"fmt"`

	// The attestation statement data sent back if attestation is requested.
	AttStatement map[string]any `json:"attStmt,omitempty"`
}
```

AttestationObject is the raw attestationObject.

Authenticators SHOULD also provide some form of attestation, if possible. If an authenticator does, the basic requirement is that the authenticator can produce, for each credential public key, an attestation statement verifiable by the WebAuthn Relying Party. Typically, this attestation statement contains a signature by an attestation private key over the attested credential public key and a challenge, as well as a certificate or similar data providing provenance information for the attestation public key, enabling the Relying Party to make a trust decision. However, if an attestation key pair is not available, then the authenticator MAY either perform self attestation of the credential public key with the corresponding credential private key, or otherwise perform no attestation. All this information is returned by authenticators any time a new public key credential is generated, in the overall form of an attestation object.

Specification: §6.5. Attestation ([https://www.w3.org/TR/webauthn/#sctn-attestation](https://www.w3.org/TR/webauthn/#sctn-attestation))

#### func (\*AttestationObject) [Verify](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/attestation.go#L131) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AttestationObject.Verify "Go to AttestationObject.Verify")

```
func (a *AttestationObject) Verify(relyingPartyID string, clientDataHash []byte, userVerificationRequired bool, userPresenceRequired bool, mds metadata.Provider, credParams []CredentialParameter) (err error)
```

Verify performs Steps 13 through 19 of registration verification.

Steps 13 through 15 are verified against the auth data. These steps are identical to 15 through 18 for assertion so we handle them with AuthData.

#### func (\*AttestationObject) [VerifyAttestation](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/attestation.go#L164) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AttestationObject.VerifyAttestation "Go to AttestationObject.VerifyAttestation")added in **v0.11.0**

```
func (a *AttestationObject) VerifyAttestation(clientDataHash []byte, mds metadata.Provider) (err error)
```

VerifyAttestation only verifies the attestation object excluding the AuthData values. If you wish to also verify the AuthData values you should use [Verify].

#### type [AttestedCredentialData](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/authenticator.go#L52) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AttestedCredentialData "Go to AttestedCredentialData")

```
type AttestedCredentialData struct {
	AAGUID       []byte `json:"aaguid"`
	CredentialID []byte `json:"credential_id"`

	// The raw credential public key bytes received from the attestation data. This is the CBOR representation of the
	// credentials public key.
	CredentialPublicKey []byte `json:"public_key"`
}
```

#### type [AuthenticationExtensions](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/options.go#L105) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticationExtensions "Go to AuthenticationExtensions")

```
type AuthenticationExtensions map[string]any
```

AuthenticationExtensions represents the AuthenticationExtensionsClientInputs IDL. This member contains additional parameters requesting additional processing by the client and authenticator.

Specification: §5.7.1. Authentication Extensions Client Inputs ([https://www.w3.org/TR/webauthn/#iface-authentication-extensions-client-inputs](https://www.w3.org/TR/webauthn/#iface-authentication-extensions-client-inputs))

#### type [AuthenticationExtensionsClientOutputs](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/extensions.go#L8) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticationExtensionsClientOutputs "Go to AuthenticationExtensionsClientOutputs")

```
type AuthenticationExtensionsClientOutputs map[string]any
```

#### type [AuthenticatorAssertionResponse](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/assertion.go#L33) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorAssertionResponse "Go to AuthenticatorAssertionResponse")

```
type AuthenticatorAssertionResponse struct {
	AuthenticatorResponse

	AuthenticatorData URLEncodedBase64 `json:"authenticatorData"`
	Signature         URLEncodedBase64 `json:"signature"`
	UserHandle        URLEncodedBase64 `json:"userHandle,omitempty"`
}
```

The AuthenticatorAssertionResponse contains the raw authenticator assertion data and is parsed into [ParsedAssertionResponse](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ParsedAssertionResponse).

#### type [AuthenticatorAttachment](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/authenticator.go#L108) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorAttachment "Go to AuthenticatorAttachment")

```
type AuthenticatorAttachment string
```

AuthenticatorAttachment represents the IDL enum of the same name, and is used as part of the Authenticator Selection Criteria.

This enumeration’s values describe authenticators' attachment modalities. Relying Parties use this to express a preferred authenticator attachment modality when calling navigator.credentials.create() to create a credential.

If this member is present, eligible authenticators are filtered to only authenticators attached with the specified §5.4.5 Authenticator Attachment Enumeration (enum AuthenticatorAttachment). The value SHOULD be a member of AuthenticatorAttachment but client platforms MUST ignore unknown values, treating an unknown value as if the member does not exist.

Specification: §5.4.4. Authenticator Selection Criteria ([https://www.w3.org/TR/webauthn/#dom-authenticatorselectioncriteria-authenticatorattachment](https://www.w3.org/TR/webauthn/#dom-authenticatorselectioncriteria-authenticatorattachment))

Specification: §5.4.5. Authenticator Attachment Enumeration ([https://www.w3.org/TR/webauthn/#enum-attachment](https://www.w3.org/TR/webauthn/#enum-attachment))

```
const (
	// Platform represents a platform authenticator is attached using a client device-specific transport, called
	// platform attachment, and is usually not removable from the client device. A public key credential bound to a
	// platform authenticator is called a platform credential.
	Platform AuthenticatorAttachment = "platform"

	// CrossPlatform represents a roaming authenticator is attached using cross-platform transports, called
	// cross-platform attachment. Authenticators of this class are removable from, and can "roam" among, client devices.
	// A public key credential bound to a roaming authenticator is called a roaming credential.
	CrossPlatform AuthenticatorAttachment = "cross-platform"
)
```

#### type [AuthenticatorAttestationResponse](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/attestation.go#L22) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorAttestationResponse "Go to AuthenticatorAttestationResponse")

```
type AuthenticatorAttestationResponse struct {
	// The byte slice of clientDataJSON, which becomes CollectedClientData.
	AuthenticatorResponse

	Transports []string `json:"transports,omitempty"`

	AuthenticatorData URLEncodedBase64 `json:"authenticatorData"`

	PublicKey URLEncodedBase64 `json:"publicKey"`

	PublicKeyAlgorithm int64 `json:"publicKeyAlgorithm"`

	// AttestationObject is the byte slice version of attestationObject.
	// This attribute contains an attestation object, which is opaque to, and
	// cryptographically protected against tampering by, the client. The
	// attestation object contains both authenticator data and an attestation
	// statement. The former contains the AAGUID, a unique credential ID, and
	// the credential public key. The contents of the attestation statement are
	// determined by the attestation statement format used by the authenticator.
	// It also contains any additional information that the Relying Party's server
	// requires to validate the attestation statement, as well as to decode and
	// validate the authenticator data along with the JSON-serialized client data.
	AttestationObject URLEncodedBase64 `json:"attestationObject"`
}
```

AuthenticatorAttestationResponse is the initial unpacked 'response' object received by the relying party. This contains the clientDataJSON object, which will be marshalled into [CollectedClientData](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#CollectedClientData), and the 'attestationObject', which contains information about the authenticator, and the newly minted public key credential. The information in both objects are used to verify the authenticity of the ceremony and new credential.

See: [https://www.w3.org/TR/webauthn/#typedefdef-publickeycredentialjson](https://www.w3.org/TR/webauthn/#typedefdef-publickeycredentialjson)

#### func (\*AuthenticatorAttestationResponse) [Parse](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/attestation.go#L94) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorAttestationResponse.Parse "Go to AuthenticatorAttestationResponse.Parse")

```
func (ccr *AuthenticatorAttestationResponse) Parse() (p *ParsedAttestationResponse, err error)
```

Parse the values returned in the authenticator response and perform attestation verification Step 8. This returns a fully decoded struct with the data put into a format that can be used to verify the user and credential that was created.

#### type [AuthenticatorData](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/authenticator.go#L44) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorData "Go to AuthenticatorData")

```
type AuthenticatorData struct {
	RPIDHash []byte                 `json:"rpid"`
	Flags    AuthenticatorFlags     `json:"flags"`
	Counter  uint32                 `json:"sign_count"`
	AttData  AttestedCredentialData `json:"att_data"`
	ExtData  []byte                 `json:"ext_data"`
}
```

AuthenticatorData represents the IDL with the same name.

The authenticator data structure encodes contextual bindings made by the authenticator. These bindings are controlled by the authenticator itself, and derive their trust from the WebAuthn Relying Party's assessment of the security properties of the authenticator. In one extreme case, the authenticator may be embedded in the client, and its bindings may be no more trustworthy than the client data. At the other extreme, the authenticator may be a discrete entity with high-security hardware and software, connected to the client over a secure channel. In both cases, the Relying Party receives the authenticator data in the same format, and uses its knowledge of the authenticator to make trust decisions.

The authenticator data has a compact but extensible encoding. This is desired since authenticators can be devices with limited capabilities and low power requirements, with much simpler software stacks than the client platform.

Specification: §6.1. Authenticator Data ([https://www.w3.org/TR/webauthn/#sctn-authenticator-data](https://www.w3.org/TR/webauthn/#sctn-authenticator-data))

#### func (\*AuthenticatorData) [Unmarshal](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/authenticator.go#L293) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorData.Unmarshal "Go to AuthenticatorData.Unmarshal")

```
func (a *AuthenticatorData) Unmarshal(rawAuthData []byte) (err error)
```

Unmarshal will take the raw Authenticator Data and marshals it into AuthenticatorData for further validation. The authenticator data has a compact but extensible encoding. This is desired since authenticators can be devices with limited capabilities and low power requirements, with much simpler software stacks than the client platform. The authenticator data structure is a byte array of 37 bytes or more, and is laid out in this table: [https://www.w3.org/TR/webauthn/#table-authData](https://www.w3.org/TR/webauthn/#table-authData)

#### func (\*AuthenticatorData) [Verify](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/authenticator.go#L392) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorData.Verify "Go to AuthenticatorData.Verify")

```
func (a *AuthenticatorData) Verify(rpIdHash []byte, appIDHash []byte, userVerificationRequired bool, userPresenceRequired bool) (err error)
```

Verify on AuthenticatorData handles Steps 13 through 15 & 17 for Registration and Steps 15 through 18 for Assertion.

#### type [AuthenticatorFlags](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/authenticator.go#L214) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorFlags "Go to AuthenticatorFlags")

```
type AuthenticatorFlags byte
```

AuthenticatorFlags A byte of information returned during during ceremonies in the authenticatorData that contains bits that give us information about the whether the user was present and/or verified during authentication, and whether there is attestation or extension data present. Bit 0 is the least significant bit.

Specification: §6.1. Authenticator Data - Flags ([https://www.w3.org/TR/webauthn/#flags](https://www.w3.org/TR/webauthn/#flags))

```
const (
	// FlagUserPresent Bit 00000001 in the byte sequence. Tells us if user is present. Also referred to as the UP flag.
	FlagUserPresent AuthenticatorFlags = 1 << iota // Referred to as UP.

	// FlagRFU1 is a reserved for future use flag.
	FlagRFU1

	// FlagUserVerified Bit 00000100 in the byte sequence. Tells us if user is verified
	// by the authenticator using a biometric or PIN. Also referred to as the UV flag.
	FlagUserVerified

	// FlagBackupEligible Bit 00001000 in the byte sequence. Tells us if a backup is eligible for device. Also referred
	// to as the BE flag.
	FlagBackupEligible // Referred to as BE.

	// FlagBackupState Bit 00010000 in the byte sequence. Tells us if a backup state for device. Also referred to as the
	// BS flag.
	FlagBackupState

	// FlagRFU2 is a reserved for future use flag.
	FlagRFU2

	// FlagAttestedCredentialData Bit 01000000 in the byte sequence. Indicates whether
	// the authenticator added attested credential data. Also referred to as the AT flag.
	FlagAttestedCredentialData

	// FlagHasExtensions Bit 10000000 in the byte sequence. Indicates if the authenticator data has extensions. Also
	// referred to as the ED flag.
	FlagHasExtensions
)
```

The bits that do not have flags are reserved for future use.

#### func (AuthenticatorFlags) [HasAttestedCredentialData](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/authenticator.go#L269) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorFlags.HasAttestedCredentialData "Go to AuthenticatorFlags.HasAttestedCredentialData")

```
func (flag AuthenticatorFlags) HasAttestedCredentialData() bool
```

HasAttestedCredentialData returns if the AT flag was set.

#### func (AuthenticatorFlags) [HasBackupEligible](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/authenticator.go#L279) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorFlags.HasBackupEligible "Go to AuthenticatorFlags.HasBackupEligible")added in **v0.6.0**

```
func (flag AuthenticatorFlags) HasBackupEligible() bool
```

HasBackupEligible returns if the BE flag was set.

#### func (AuthenticatorFlags) [HasBackupState](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/authenticator.go#L284) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorFlags.HasBackupState "Go to AuthenticatorFlags.HasBackupState")added in **v0.6.0**

```
func (flag AuthenticatorFlags) HasBackupState() bool
```

HasBackupState returns if the BS flag was set.

#### func (AuthenticatorFlags) [HasExtensions](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/authenticator.go#L274) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorFlags.HasExtensions "Go to AuthenticatorFlags.HasExtensions")

```
func (flag AuthenticatorFlags) HasExtensions() bool
```

HasExtensions returns if the ED flag was set.

#### func (AuthenticatorFlags) [HasUserPresent](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/authenticator.go#L259) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorFlags.HasUserPresent "Go to AuthenticatorFlags.HasUserPresent")added in **v0.7.2**

```
func (flag AuthenticatorFlags) HasUserPresent() bool
```

HasUserPresent returns if the UP flag was set.

#### func (AuthenticatorFlags) [HasUserVerified](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/authenticator.go#L264) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorFlags.HasUserVerified "Go to AuthenticatorFlags.HasUserVerified")added in **v0.7.2**

```
func (flag AuthenticatorFlags) HasUserVerified() bool
```

HasUserVerified returns if the UV flag was set.

#### func (AuthenticatorFlags) [UserPresent](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/authenticator.go#L249) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorFlags.UserPresent "Go to AuthenticatorFlags.UserPresent")

```
func (flag AuthenticatorFlags) UserPresent() bool
```

UserPresent returns if the UP flag was set.

#### func (AuthenticatorFlags) [UserVerified](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/authenticator.go#L254) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorFlags.UserVerified "Go to AuthenticatorFlags.UserVerified")

```
func (flag AuthenticatorFlags) UserVerified() bool
```

UserVerified returns if the UV flag was set.

#### type [AuthenticatorResponse](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/authenticator.go#L23) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorResponse "Go to AuthenticatorResponse")

```
type AuthenticatorResponse struct {
	// From the spec https://www.w3.org/TR/webauthn/#dom-authenticatorresponse-clientdatajson
	// This attribute contains a JSON serialization of the client data passed to the authenticator
	// by the client in its call to either create() or get().
	ClientDataJSON URLEncodedBase64 `json:"clientDataJSON"`
}
```

AuthenticatorResponse represents the IDL with the same name.

Authenticators respond to Relying Party requests by returning an object derived from the AuthenticatorResponse interface

Specification: §5.2. Authenticator Responses ([https://www.w3.org/TR/webauthn/#iface-authenticatorresponse](https://www.w3.org/TR/webauthn/#iface-authenticatorresponse))

#### type [AuthenticatorSelection](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/options.go#L113) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorSelection "Go to AuthenticatorSelection")

```
type AuthenticatorSelection struct {
	// AuthenticatorAttachment If this member is present, eligible authenticators are filtered to only
	// authenticators attached with the specified AuthenticatorAttachment enum.
	AuthenticatorAttachment AuthenticatorAttachment `json:"authenticatorAttachment,omitempty"`

	// RequireResidentKey this member describes the Relying Party's requirements regarding resident
	// credentials. If the parameter is set to true, the authenticator MUST create a client-side-resident
	// public key credential source when creating a public key credential.
	RequireResidentKey *bool `json:"requireResidentKey,omitempty"`

	// ResidentKey this member describes the Relying Party's requirements regarding resident
	// credentials per Webauthn Level 2.
	ResidentKey ResidentKeyRequirement `json:"residentKey,omitempty"`

	// UserVerification This member describes the Relying Party's requirements regarding user verification for
	// the create() operation. Eligible authenticators are filtered to only those capable of satisfying this
	// requirement.
	UserVerification UserVerificationRequirement `json:"userVerification,omitempty"`
}
```

AuthenticatorSelection represents the AuthenticatorSelectionCriteria IDL.

WebAuthn Relying Parties may use the AuthenticatorSelectionCriteria dictionary to specify their requirements regarding authenticator attributes.

Specification: §5.4.4. Authenticator Selection Criteria ([https://www.w3.org/TR/webauthn/#dictionary-authenticatorSelection](https://www.w3.org/TR/webauthn/#dictionary-authenticatorSelection))

#### type [AuthenticatorTransport](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/authenticator.go#L160) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorTransport "Go to AuthenticatorTransport")

```
type AuthenticatorTransport string
```

AuthenticatorTransport represents the IDL enum with the same name.

Authenticators may implement various transports for communicating with clients. This enumeration defines hints as to how clients might communicate with a particular authenticator in order to obtain an assertion for a specific credential. Note that these hints represent the WebAuthn Relying Party's best belief as to how an authenticator may be reached. A Relying Party will typically learn of the supported transports for a public key credential via getTransports().

Specification: §5.8.4. Authenticator Transport Enumeration ([https://www.w3.org/TR/webauthn/#enumdef-authenticatortransport](https://www.w3.org/TR/webauthn/#enumdef-authenticatortransport))

```
const (
	// USB indicates the respective authenticator can be contacted over removable USB.
	USB AuthenticatorTransport = "usb"

	// NFC indicates the respective authenticator can be contacted over Near Field Communication (NFC).
	NFC AuthenticatorTransport = "nfc"

	// BLE indicates the respective authenticator can be contacted over Bluetooth Smart (Bluetooth Low Energy / BLE).
	BLE AuthenticatorTransport = "ble"

	// SmartCard indicates the respective authenticator can be contacted over ISO/IEC 7816 smart card with contacts.
	//
	// WebAuthn Level 3.
	SmartCard AuthenticatorTransport = "smart-card"

	// Hybrid indicates the respective authenticator can be contacted using a combination of (often separate)
	// data-transport and proximity mechanisms. This supports, for example, authentication on a desktop computer using
	// a smartphone.
	//
	// WebAuthn Level 3.
	Hybrid AuthenticatorTransport = "hybrid"

	// Internal indicates the respective authenticator is contacted using a client device-specific transport, i.e., it
	// is a platform authenticator. These authenticators are not removable from the client device.
	Internal AuthenticatorTransport = "internal"
)
```

#### type [CeremonyType](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/client.go#L31) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#CeremonyType "Go to CeremonyType")

```
type CeremonyType string
```

```
const (
	CreateCeremony CeremonyType = "webauthn.create"
	AssertCeremony CeremonyType = "webauthn.get"
)
```

#### type [CollectedClientData](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/client.go#L15) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#CollectedClientData "Go to CollectedClientData")

```
type CollectedClientData struct {
	// Type the string "webauthn.create" when creating new credentials,
	// and "webauthn.get" when getting an assertion from an existing credential. The
	// purpose of this member is to prevent certain types of signature confusion attacks
	// (where an attacker substitutes one legitimate signature for another).
	Type         CeremonyType  `json:"type"`
	Challenge    string        `json:"challenge"`
	Origin       string        `json:"origin"`
	TopOrigin    string        `json:"topOrigin,omitempty"`
	CrossOrigin  bool          `json:"crossOrigin,omitempty"`
	TokenBinding *TokenBinding `json:"tokenBinding,omitempty"`

	// Chromium (Chrome) returns a hint sometimes about how to handle clientDataJSON in a safe manner.
	Hint string `json:"new_keys_may_be_added_here,omitempty"`
}
```

CollectedClientData represents the contextual bindings of both the WebAuthn Relying Party and the client. It is a key-value mapping whose keys are strings. Values can be any type that has a valid encoding in JSON. Its structure is defined by the following Web IDL.

Specification: §5.8.1. Client Data Used in WebAuthn Signatures ([https://www.w3.org/TR/webauthn/#dictdef-collectedclientdata](https://www.w3.org/TR/webauthn/#dictdef-collectedclientdata))

#### func (\*CollectedClientData) [Verify](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/client.go#L87) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#CollectedClientData.Verify "Go to CollectedClientData.Verify")

```
func (c *CollectedClientData) Verify(storedChallenge string, ceremony CeremonyType, rpOrigins, rpTopOrigins []string, rpTopOriginsVerify TopOriginVerificationMode) (err error)
```

Verify handles steps 3 through 6 of verifying the registering client data of a new credential and steps 7 through 10 of verifying an authentication assertion See [https://www.w3.org/TR/webauthn/#registering-a-new-credential](https://www.w3.org/TR/webauthn/#registering-a-new-credential) and [https://www.w3.org/TR/webauthn/#verifying-assertion](https://www.w3.org/TR/webauthn/#verifying-assertion)

Note: the rpTopOriginsVerify parameter does not accept the TopOriginVerificationMode value of TopOriginDefaultVerificationMode as it's expected this value is updated by the config validation process.

#### type [ConveyancePreference](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/options.go#L139) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ConveyancePreference "Go to ConveyancePreference")

```
type ConveyancePreference string
```

ConveyancePreference is the type representing the AttestationConveyancePreference IDL.

WebAuthn Relying Parties may use AttestationConveyancePreference to specify their preference regarding attestation conveyance during credential generation.

Specification: §5.4.7. Attestation Conveyance Preference Enumeration ([https://www.w3.org/TR/webauthn/#enum-attestation-convey](https://www.w3.org/TR/webauthn/#enum-attestation-convey))

```
const (
	// PreferNoAttestation is a ConveyancePreference value.
	//
	// This value indicates that the Relying Party is not interested in authenticator attestation. For example, in order
	// to potentially avoid having to obtain user consent to relay identifying information to the Relying Party, or to
	// save a round trip to an Attestation CA or Anonymization CA.
	//
	// This is the default value.
	//
	// Specification: §5.4.7. Attestation Conveyance Preference Enumeration (https://www.w3.org/TR/webauthn/#dom-attestationconveyancepreference-none)
	PreferNoAttestation ConveyancePreference = "none"

	// PreferIndirectAttestation is a ConveyancePreference value.
	//
	// This value indicates that the Relying Party prefers an attestation conveyance yielding verifiable attestation
	// statements, but allows the client to decide how to obtain such attestation statements. The client MAY replace the
	// authenticator-generated attestation statements with attestation statements generated by an Anonymization CA, in
	// order to protect the user’s privacy, or to assist Relying Parties with attestation verification in a
	// heterogeneous ecosystem.
	//
	// Note: There is no guarantee that the Relying Party will obtain a verifiable attestation statement in this case.
	// For example, in the case that the authenticator employs self attestation.
	//
	// Specification: §5.4.7. Attestation Conveyance Preference Enumeration (https://www.w3.org/TR/webauthn/#dom-attestationconveyancepreference-indirect)
	PreferIndirectAttestation ConveyancePreference = "indirect"

	// PreferDirectAttestation is a ConveyancePreference value.
	//
	// This value indicates that the Relying Party wants to receive the attestation statement as generated by the
	// authenticator.
	//
	// Specification: §5.4.7. Attestation Conveyance Preference Enumeration (https://www.w3.org/TR/webauthn/#dom-attestationconveyancepreference-direct)
	PreferDirectAttestation ConveyancePreference = "direct"

	// PreferEnterpriseAttestation is a ConveyancePreference value.
	//
	// This value indicates that the Relying Party wants to receive an attestation statement that may include uniquely
	// identifying information. This is intended for controlled deployments within an enterprise where the organization
	// wishes to tie registrations to specific authenticators. User agents MUST NOT provide such an attestation unless
	// the user agent or authenticator configuration permits it for the requested RP ID.
	//
	// If permitted, the user agent SHOULD signal to the authenticator (at invocation time) that enterprise
	// attestation is requested, and convey the resulting AAGUID and attestation statement, unaltered, to the Relying
	// Party.
	//
	// Specification: §5.4.7. Attestation Conveyance Preference Enumeration (https://www.w3.org/TR/webauthn/#dom-attestationconveyancepreference-enterprise)
	PreferEnterpriseAttestation ConveyancePreference = "enterprise"
)
```

#### type [Credential](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/credential.go#L16) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#Credential "Go to Credential")

```
type Credential struct {
	// ID is The credential’s identifier. The requirements for the
	// identifier are distinct for each type of credential. It might
	// represent a username for username/password tuples, for example.
	ID string `json:"id"`
	// Type is the value of the object’s interface object's [[type]] slot,
	// which specifies the credential type represented by this object.
	// This should be type "public-key" for Webauthn credentials.
	Type string `json:"type"`
}
```

Credential is the basic credential type from the Credential Management specification that is inherited by WebAuthn's PublicKeyCredential type.

Specification: Credential Management §2.2. The Credential Interface ([https://www.w3.org/TR/credential-management/#credential](https://www.w3.org/TR/credential-management/#credential))

#### type [CredentialAssertion](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/options.go#L12) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#CredentialAssertion "Go to CredentialAssertion")

```
type CredentialAssertion struct {
	Response  PublicKeyCredentialRequestOptions `json:"publicKey"`
	Mediation CredentialMediationRequirement    `json:"mediation,omitempty"`
}
```

#### type [CredentialAssertionResponse](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/assertion.go#L16) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#CredentialAssertionResponse "Go to CredentialAssertionResponse")

```
type CredentialAssertionResponse struct {
	PublicKeyCredential

	AssertionResponse AuthenticatorAssertionResponse `json:"response"`
}
```

The CredentialAssertionResponse is the raw response returned to the Relying Party from an authenticator when we request a credential for login/assertion.

#### func (CredentialAssertionResponse) [Parse](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/assertion.go#L94) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#CredentialAssertionResponse.Parse "Go to CredentialAssertionResponse.Parse")added in **v0.8.2**

```
func (car CredentialAssertionResponse) Parse() (par *ParsedCredentialAssertionData, err error)
```

Parse validates and parses the [CredentialAssertionResponse](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#CredentialAssertionResponse) into a [ParseCredentialCreationResponseBody](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ParseCredentialCreationResponseBody). This receiver is unlikely to be expressly guaranteed under the versioning policy. Users looking for this guarantee should see [ParseCredentialRequestResponseBody](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ParseCredentialRequestResponseBody) instead, and this receiver should only be used if that function is inadequate for their use case.

#### type [CredentialCreation](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/options.go#L7) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#CredentialCreation "Go to CredentialCreation")

```
type CredentialCreation struct {
	Response  PublicKeyCredentialCreationOptions `json:"publicKey"`
	Mediation CredentialMediationRequirement     `json:"mediation,omitempty"`
}
```

#### type [CredentialCreationResponse](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/credential.go#L50) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#CredentialCreationResponse "Go to CredentialCreationResponse")

```
type CredentialCreationResponse struct {
	PublicKeyCredential

	AttestationResponse AuthenticatorAttestationResponse `json:"response"`
}
```

#### func (CredentialCreationResponse) [Parse](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/credential.go#L106) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#CredentialCreationResponse.Parse "Go to CredentialCreationResponse.Parse")added in **v0.8.2**

```
func (ccr CredentialCreationResponse) Parse() (pcc *ParsedCredentialCreationData, err error)
```

Parse validates and parses the CredentialCreationResponse into a ParsedCredentialCreationData. This receiver is unlikely to be expressly guaranteed under the versioning policy. Users looking for this guarantee should see ParseCredentialCreationResponseBody instead, and this receiver should only be used if that function is inadequate for their use case.

#### type [CredentialDescriptor](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/options.go#L62) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#CredentialDescriptor "Go to CredentialDescriptor")

```
type CredentialDescriptor struct {
	// The valid credential types.
	Type CredentialType `json:"type"`

	// CredentialID The ID of a credential to allow/disallow.
	CredentialID URLEncodedBase64 `json:"id"`

	// The authenticator transports that can be used.
	Transport []AuthenticatorTransport `json:"transports,omitempty"`

	// The AttestationType from the Credential. Used internally only.
	AttestationType string `json:"-"`
}
```

CredentialDescriptor represents the PublicKeyCredentialDescriptor IDL.

This dictionary contains the attributes that are specified by a caller when referring to a public key credential as an input parameter to the create() or get() methods. It mirrors the fields of the PublicKeyCredential object returned by the latter methods.

Specification: §5.10.3. Credential Descriptor ([https://www.w3.org/TR/webauthn/#credential-dictionary](https://www.w3.org/TR/webauthn/#credential-dictionary))

#### type [CredentialEntity](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/entities.go#L7) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#CredentialEntity "Go to CredentialEntity")

```
type CredentialEntity struct {
	// A human-palatable name for the entity. Its function depends on what the PublicKeyCredentialEntity represents:
	//
	// When inherited by PublicKeyCredentialRpEntity it is a human-palatable identifier for the Relying Party,
	// intended only for display. For example, "ACME Corporation", "Wonderful Widgets, Inc." or "ОАО Примертех".
	//
	// When inherited by PublicKeyCredentialUserEntity, it is a human-palatable identifier for a user account. It is
	// intended only for display, i.e., aiding the user in determining the difference between user accounts with similar
	// displayNames. For example, "alexm", "alex.p.mueller@example.com" or "+14255551234".
	Name string `json:"name"`
}
```

CredentialEntity represents the PublicKeyCredentialEntity IDL and it describes a user account, or a WebAuthn Relying Party with which a public key credential is associated.

Specification: §5.4.1. Public Key Entity Description ([https://www.w3.org/TR/webauthn/#dictionary-pkcredentialentity](https://www.w3.org/TR/webauthn/#dictionary-pkcredentialentity))

#### type [CredentialMediationRequirement](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/authenticator.go#L66) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#CredentialMediationRequirement "Go to CredentialMediationRequirement")added in **v0.12.0**

```
type CredentialMediationRequirement string
```

CredentialMediationRequirement represents mediation requirements for clients. When making a request via get(options) or create(options), developers can set a case-by-case requirement for user mediation by choosing the appropriate CredentialMediationRequirement enum value.

See [https://www.w3.org/TR/credential-management-1/#mediation-requirements](https://www.w3.org/TR/credential-management-1/#mediation-requirements)

```
const (
	// MediationDefault lets the browser choose the mediation flow completely as if it wasn't specified at all.
	MediationDefault CredentialMediationRequirement = ""

	// MediationSilent indicates user mediation is suppressed for the given operation. If the operation can be performed
	// without user involvement, wonderful. If user involvement is necessary, then the operation will return null rather
	// than involving the user.
	MediationSilent CredentialMediationRequirement = "silent"

	// MediationOptional indicates if credentials can be handed over for a given operation without user mediation, they
	// will be. If user mediation is required, then the user agent will involve the user in the decision.
	MediationOptional CredentialMediationRequirement = "optional"

	// MediationConditional indicates for get(), discovered credentials are presented to the user in a non-modal dialog
	// along with an indication of the origin which is requesting credentials. If the user makes a gesture outside of
	// the dialog, the dialog closes without resolving or rejecting the Promise returned by the get() method and without
	// causing a user-visible error condition. If the user makes a gesture that selects a credential, that credential is
	// returned to the caller. The prevent silent access flag is treated as being true regardless of its actual value:
	// the conditional behavior always involves user mediation of some sort if applicable credentials are discovered.
	MediationConditional CredentialMediationRequirement = "conditional"

	// MediationRequired indicates the user agent will not hand over credentials without user mediation, even if the
	// prevent silent access flag is unset for an origin.
	MediationRequired CredentialMediationRequirement = "required"
)
```

#### type [CredentialParameter](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/options.go#L78) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#CredentialParameter "Go to CredentialParameter")

```
type CredentialParameter struct {
	Type      CredentialType                       `json:"type"`
	Algorithm webauthncose.COSEAlgorithmIdentifier `json:"alg"`
}
```

CredentialParameter is the credential type and algorithm that the relying party wants the authenticator to create.

#### type [CredentialType](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/options.go#L94) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#CredentialType "Go to CredentialType")

```
type CredentialType string
```

CredentialType represents the PublicKeyCredentialType IDL and is used with the CredentialDescriptor IDL.

This enumeration defines the valid credential types. It is an extension point; values can be added to it in the future, as more credential types are defined. The values of this enumeration are used for versioning the Authentication Assertion and attestation structures according to the type of the authenticator.

Currently one credential type is defined, namely "public-key".

Specification: §5.8.2. Credential Type Enumeration ([https://www.w3.org/TR/webauthn/#enumdef-publickeycredentialtype](https://www.w3.org/TR/webauthn/#enumdef-publickeycredentialtype))

Specification: §5.8.3. Credential Descriptor ([https://www.w3.org/TR/webauthn/#dictionary-credential-descriptor](https://www.w3.org/TR/webauthn/#dictionary-credential-descriptor))

```
const (
	// PublicKeyCredentialType - Currently one credential type is defined, namely "public-key".
	PublicKeyCredentialType CredentialType = "public-key"
)
```

#### type [Error](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/errors.go#L3) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#Error "Go to Error")

```
type Error struct {
	// Short name for the type of error that has occurred.
	Type string `json:"type"`

	// Additional details about the error.
	Details string `json:"error"`

	// Information to help debug the error.
	DevInfo string `json:"debug"`

	// Inner error.
	Err error `json:"-"`
}
```

#### func [ValidateMetadata](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/metadata.go#L13) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ValidateMetadata "Go to ValidateMetadata")added in **v0.11.0**

```
func ValidateMetadata(ctx context.Context, mds metadata.Provider, aaguid uuid.UUID, attestationType, attestationFormat string, x5cs []any) (protoErr *Error)
```

#### func (\*Error) [Error](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/errors.go#L17) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#Error.Error "Go to Error.Error")

```
func (e *Error) Error() string
```

#### func (\*Error) [Unwrap](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/errors.go#L21) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#Error.Unwrap "Go to Error.Unwrap")added in **v0.12.0**

```
func (e *Error) Unwrap() error
```

#### func (\*Error) [WithDetails](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/errors.go#L25) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#Error.WithDetails "Go to Error.WithDetails")

```
func (e *Error) WithDetails(details string) *Error
```

#### func (\*Error) [WithError](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/errors.go#L39) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#Error.WithError "Go to Error.WithError")added in **v0.12.0**

```
func (e *Error) WithError(err error) *Error
```

#### func (\*Error) [WithInfo](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/errors.go#L32) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#Error.WithInfo "Go to Error.WithInfo")

```
func (e *Error) WithInfo(info string) *Error
```

#### type [Extensions](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/options.go#L266) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#Extensions "Go to Extensions")

```
type Extensions any
```

#### type [ParsedAssertionResponse](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/assertion.go#L42) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ParsedAssertionResponse "Go to ParsedAssertionResponse")

```
type ParsedAssertionResponse struct {
	CollectedClientData CollectedClientData
	AuthenticatorData   AuthenticatorData
	Signature           []byte
	UserHandle          []byte
}
```

ParsedAssertionResponse is the parsed form of [AuthenticatorAssertionResponse](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorAssertionResponse).

#### type [ParsedAttestationResponse](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/attestation.go#L48) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ParsedAttestationResponse "Go to ParsedAttestationResponse")

```
type ParsedAttestationResponse struct {
	CollectedClientData CollectedClientData
	AttestationObject   AttestationObject
	Transports          []AuthenticatorTransport
}
```

ParsedAttestationResponse is the parsed version of [AuthenticatorAttestationResponse](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#AuthenticatorAttestationResponse).

#### type [ParsedCredential](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/credential.go#L29) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ParsedCredential "Go to ParsedCredential")

```
type ParsedCredential struct {
	ID   string `cbor:"id"`
	Type string `cbor:"type"`
}
```

ParsedCredential is the parsed PublicKeyCredential interface, inherits from Credential, and contains the attributes that are returned to the caller when a new credential is created, or a new assertion is requested.

#### type [ParsedCredentialAssertionData](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/assertion.go#L24) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ParsedCredentialAssertionData "Go to ParsedCredentialAssertionData")

```
type ParsedCredentialAssertionData struct {
	ParsedPublicKeyCredential

	Response ParsedAssertionResponse
	Raw      CredentialAssertionResponse
}
```

The ParsedCredentialAssertionData is the parsed [CredentialAssertionResponse](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#CredentialAssertionResponse) that has been marshalled into a format that allows us to verify the client and authenticator data inside the response.

#### func [ParseCredentialRequestResponse](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/assertion.go#L52) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ParseCredentialRequestResponse "Go to ParseCredentialRequestResponse")

```
func ParseCredentialRequestResponse(response *http.Request) (*ParsedCredentialAssertionData, error)
```

ParseCredentialRequestResponse parses the credential request response into a format that is either required by the specification or makes the assertion verification steps easier to complete. This takes a [\*http.Request](https://pkg.go.dev/net/http#Request) that contains the assertion response data in a raw, mostly base64 encoded format, and parses the data into manageable structures.

#### func [ParseCredentialRequestResponseBody](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/assertion.go#L68) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ParseCredentialRequestResponseBody "Go to ParseCredentialRequestResponseBody")

```
func ParseCredentialRequestResponseBody(body io.Reader) (par *ParsedCredentialAssertionData, err error)
```

ParseCredentialRequestResponseBody parses the credential request response into a format that is either required by the specification or makes the assertion verification steps easier to complete. This takes an [io.Reader](https://pkg.go.dev/io#Reader) that contains the assertion response data in a raw, mostly base64 encoded format, and parses the data into manageable structures.

#### func [ParseCredentialRequestResponseBytes](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/assertion.go#L80) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ParseCredentialRequestResponseBytes "Go to ParseCredentialRequestResponseBytes")added in **v0.11.0**

```
func ParseCredentialRequestResponseBytes(data []byte) (par *ParsedCredentialAssertionData, err error)
```

ParseCredentialRequestResponseBytes is an alternative version of [ParseCredentialRequestResponseBody](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ParseCredentialRequestResponseBody) that just takes a byte slice.

#### func (\*ParsedCredentialAssertionData) [Verify](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/assertion.go#L142) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ParsedCredentialAssertionData.Verify "Go to ParsedCredentialAssertionData.Verify")

```
func (p *ParsedCredentialAssertionData) Verify(storedChallenge string, relyingPartyID string, rpOrigins, rpTopOrigins []string, rpTopOriginsVerify TopOriginVerificationMode, appID string, verifyUser bool, verifyUserPresence bool, credentialBytes []byte) error
```

Verify the remaining elements of the assertion data by following the steps outlined in the referenced specification documentation. It's important to note that the credentialBytes field is the CBOR representation of the credential.

Specification: §7.2 Verifying an Authentication Assertion ([https://www.w3.org/TR/webauthn/#sctn-verifying-assertion](https://www.w3.org/TR/webauthn/#sctn-verifying-assertion))

#### type [ParsedCredentialCreationData](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/credential.go#L56) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ParsedCredentialCreationData "Go to ParsedCredentialCreationData")

```
type ParsedCredentialCreationData struct {
	ParsedPublicKeyCredential

	Response ParsedAttestationResponse
	Raw      CredentialCreationResponse
}
```

#### func [ParseCredentialCreationResponse](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/credential.go#L65) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ParseCredentialCreationResponse "Go to ParseCredentialCreationResponse")

```
func ParseCredentialCreationResponse(request *http.Request) (*ParsedCredentialCreationData, error)
```

ParseCredentialCreationResponse is a non-agnostic function for parsing a registration response from the http library from stdlib. It handles some standard cleanup operations.

#### func [ParseCredentialCreationResponseBody](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/credential.go#L80) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ParseCredentialCreationResponseBody "Go to ParseCredentialCreationResponseBody")

```
func ParseCredentialCreationResponseBody(body io.Reader) (pcc *ParsedCredentialCreationData, err error)
```

ParseCredentialCreationResponseBody is an agnostic version of ParseCredentialCreationResponse. Implementers are therefore responsible for managing cleanup.

#### func [ParseCredentialCreationResponseBytes](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/credential.go#L92) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ParseCredentialCreationResponseBytes "Go to ParseCredentialCreationResponseBytes")added in **v0.11.0**

```
func ParseCredentialCreationResponseBytes(data []byte) (pcc *ParsedCredentialCreationData, err error)
```

ParseCredentialCreationResponseBytes is an alternative version of ParseCredentialCreationResponseBody that just takes a byte slice.

#### func (\*ParsedCredentialCreationData) [Verify](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/credential.go#L150) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ParsedCredentialCreationData.Verify "Go to ParsedCredentialCreationData.Verify")

```
func (pcc *ParsedCredentialCreationData) Verify(storedChallenge string, verifyUser bool, verifyUserPresence bool, relyingPartyID string, rpOrigins, rpTopOrigins []string, rpTopOriginsVerify TopOriginVerificationMode, mds metadata.Provider, credParams []CredentialParameter) (clientDataHash []byte, err error)
```

Verify the Client and Attestation data.

Specification: §7.1. Registering a New Credential ([https://www.w3.org/TR/webauthn/#sctn-registering-a-new-credential](https://www.w3.org/TR/webauthn/#sctn-registering-a-new-credential))

#### type [ParsedPublicKeyCredential](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/credential.go#L42) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ParsedPublicKeyCredential "Go to ParsedPublicKeyCredential")

```
type ParsedPublicKeyCredential struct {
	ParsedCredential

	RawID                   []byte                                `json:"rawId"`
	ClientExtensionResults  AuthenticationExtensionsClientOutputs `json:"clientExtensionResults,omitempty"`
	AuthenticatorAttachment AuthenticatorAttachment               `json:"authenticatorAttachment,omitempty"`
}
```

#### func (ParsedPublicKeyCredential) [GetAppID](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/credential.go#L220) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ParsedPublicKeyCredential.GetAppID "Go to ParsedPublicKeyCredential.GetAppID")

```
func (ppkc ParsedPublicKeyCredential) GetAppID(authExt AuthenticationExtensions, credentialAttestationType string) (appID string, err error)
```

GetAppID takes a AuthenticationExtensions object or nil. It then performs the following checks in order:

1. Check that the Session Data's AuthenticationExtensions has been provided and if it hasn't return an error. 2. Check that the AuthenticationExtensionsClientOutputs contains the extensions output and return an empty string if it doesn't. 3. Check that the Credential AttestationType is `fido-u2f` and return an empty string if it isn't. 4. Check that the AuthenticationExtensionsClientOutputs contains the appid key and if it doesn't return an empty string. 5. Check that the AuthenticationExtensionsClientOutputs appid is a bool and if it isn't return an error. 6. Check that the appid output is true and if it isn't return an empty string. 7. Check that the Session Data has an appid extension defined and if it doesn't return an error. 8. Check that the appid extension in Session Data is a string and if it isn't return an error. 9. Return the appid extension value from the Session data.

#### type [PublicKeyCredential](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/credential.go#L34) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#PublicKeyCredential "Go to PublicKeyCredential")

```
type PublicKeyCredential struct {
	Credential

	RawID                   URLEncodedBase64                      `json:"rawId"`
	ClientExtensionResults  AuthenticationExtensionsClientOutputs `json:"clientExtensionResults,omitempty"`
	AuthenticatorAttachment string                                `json:"authenticatorAttachment,omitempty"`
}
```

#### type [PublicKeyCredentialCreationOptions](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/options.go#L25) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#PublicKeyCredentialCreationOptions "Go to PublicKeyCredentialCreationOptions")

```
type PublicKeyCredentialCreationOptions struct {
	RelyingParty           RelyingPartyEntity         `json:"rp"`
	User                   UserEntity                 `json:"user"`
	Challenge              URLEncodedBase64           `json:"challenge"`
	Parameters             []CredentialParameter      `json:"pubKeyCredParams,omitempty"`
	Timeout                int                        `json:"timeout,omitempty"`
	CredentialExcludeList  []CredentialDescriptor     `json:"excludeCredentials,omitempty"`
	AuthenticatorSelection AuthenticatorSelection     `json:"authenticatorSelection,omitempty"`
	Hints                  []PublicKeyCredentialHints `json:"hints,omitempty"`
	Attestation            ConveyancePreference       `json:"attestation,omitempty"`
	AttestationFormats     []AttestationFormat        `json:"attestationFormats,omitempty"`
	Extensions             AuthenticationExtensions   `json:"extensions,omitempty"`
}
```

PublicKeyCredentialCreationOptions represents the IDL of the same name.

In order to create a Credential via create(), the caller specifies a few parameters in a PublicKeyCredentialCreationOptions object.

WebAuthn Level 3: hints,attestationFormats.

Specification: §5.4. Options for Credential Creation ([https://www.w3.org/TR/webauthn/#dictionary-makecredentialoptions](https://www.w3.org/TR/webauthn/#dictionary-makecredentialoptions))

#### type [PublicKeyCredentialHints](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/options.go#L227) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#PublicKeyCredentialHints "Go to PublicKeyCredentialHints")added in **v0.11.0**

```
type PublicKeyCredentialHints string
```

```
const (
	// PublicKeyCredentialHintSecurityKey is a PublicKeyCredentialHint that indicates that the Relying Party believes
	// that users will satisfy this request with a physical security key. For example, an enterprise Relying Party may
	// set this hint if they have issued security keys to their employees and will only accept those authenticators for
	// registration and authentication.
	//
	// For compatibility with older user agents, when this hint is used in PublicKeyCredentialCreationOptions, the
	// authenticatorAttachment SHOULD be set to cross-platform.
	PublicKeyCredentialHintSecurityKey PublicKeyCredentialHints = "security-key"

	// PublicKeyCredentialHintClientDevice is a PublicKeyCredentialHint that indicates that the Relying Party believes
	// that users will satisfy this request with a platform authenticator attached to the client device.
	//
	// For compatibility with older user agents, when this hint is used in PublicKeyCredentialCreationOptions, the
	// authenticatorAttachment SHOULD be set to platform.
	PublicKeyCredentialHintClientDevice PublicKeyCredentialHints = "client-device"

	// PublicKeyCredentialHintHybrid is a PublicKeyCredentialHint that indicates that the Relying Party believes that
	// users will satisfy this request with general-purpose authenticators such as smartphones. For example, a consumer
	// Relying Party may believe that only a small fraction of their customers possesses dedicated security keys. This
	// option also implies that the local platform authenticator should not be promoted in the UI.
	//
	// For compatibility with older user agents, when this hint is used in PublicKeyCredentialCreationOptions, the
	// authenticatorAttachment SHOULD be set to cross-platform.
	PublicKeyCredentialHintHybrid PublicKeyCredentialHints = "hybrid"
)
```

#### type [PublicKeyCredentialRequestOptions](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/options.go#L45) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#PublicKeyCredentialRequestOptions "Go to PublicKeyCredentialRequestOptions")

```
type PublicKeyCredentialRequestOptions struct {
	Challenge          URLEncodedBase64            `json:"challenge"`
	Timeout            int                         `json:"timeout,omitempty"`
	RelyingPartyID     string                      `json:"rpId,omitempty"`
	AllowedCredentials []CredentialDescriptor      `json:"allowCredentials,omitempty"`
	UserVerification   UserVerificationRequirement `json:"userVerification,omitempty"`
	Hints              []PublicKeyCredentialHints  `json:"hints,omitempty"`
	Extensions         AuthenticationExtensions    `json:"extensions,omitempty"`
}
```

The PublicKeyCredentialRequestOptions dictionary supplies get() with the data it needs to generate an assertion. Its challenge member MUST be present, while its other members are OPTIONAL.

WebAuthn Level 3: hints.

Specification: §5.5. Options for Assertion Generation ([https://www.w3.org/TR/webauthn/#dictionary-assertion-options](https://www.w3.org/TR/webauthn/#dictionary-assertion-options))

#### func (\*PublicKeyCredentialRequestOptions) [GetAllowedCredentialIDs](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/options.go#L256) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#PublicKeyCredentialRequestOptions.GetAllowedCredentialIDs "Go to PublicKeyCredentialRequestOptions.GetAllowedCredentialIDs")

```
func (a *PublicKeyCredentialRequestOptions) GetAllowedCredentialIDs() [][]byte
```

#### type [RelyingPartyEntity](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/entities.go#L23) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#RelyingPartyEntity "Go to RelyingPartyEntity")

```
type RelyingPartyEntity struct {
	CredentialEntity

	// A unique identifier for the Relying Party entity, which sets the RP ID.
	ID string `json:"id"`
}
```

The RelyingPartyEntity represents the PublicKeyCredentialRpEntity IDL and is used to supply additional Relying Party attributes when creating a new credential.

Specification: §5.4.2. Relying Party Parameters for Credential Generation ([https://www.w3.org/TR/webauthn/#dictionary-rp-credential-params](https://www.w3.org/TR/webauthn/#dictionary-rp-credential-params))

#### type [ResidentKeyRequirement](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/authenticator.go#L136) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ResidentKeyRequirement "Go to ResidentKeyRequirement")added in **v0.2.0**

```
type ResidentKeyRequirement string
```

ResidentKeyRequirement represents the IDL of the same name.

This enumeration’s values describe the Relying Party's requirements for client-side discoverable credentials (formerly known as resident credentials or resident keys).

Specifies the extent to which the Relying Party desires to create a client-side discoverable credential. For historical reasons the naming retains the deprecated “resident” terminology. The value SHOULD be a member of ResidentKeyRequirement but client platforms MUST ignore unknown values, treating an unknown value as if the member does not exist. If no value is given then the effective value is required if requireResidentKey is true or discouraged if it is false or absent.

Specification: §5.4.4. Authenticator Selection Criteria ([https://www.w3.org/TR/webauthn/#dom-authenticatorselectioncriteria-residentkey](https://www.w3.org/TR/webauthn/#dom-authenticatorselectioncriteria-residentkey))

Specification: §5.4.6. Resident Key Requirement Enumeration ([https://www.w3.org/TR/webauthn/#enumdef-residentkeyrequirement](https://www.w3.org/TR/webauthn/#enumdef-residentkeyrequirement))

```
const (
	// ResidentKeyRequirementDiscouraged indicates the Relying Party prefers creating a server-side credential, but will
	// accept a client-side discoverable credential. This is the default.
	ResidentKeyRequirementDiscouraged ResidentKeyRequirement = "discouraged"

	// ResidentKeyRequirementPreferred indicates to the client we would prefer a discoverable credential.
	ResidentKeyRequirementPreferred ResidentKeyRequirement = "preferred"

	// ResidentKeyRequirementRequired indicates the Relying Party requires a client-side discoverable credential, and is
	// prepared to receive an error if a client-side discoverable credential cannot be created.
	ResidentKeyRequirementRequired ResidentKeyRequirement = "required"
)
```

#### type [SafetyNetResponse](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/attestation_safetynet.go#L173) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#SafetyNetResponse "Go to SafetyNetResponse")

```
type SafetyNetResponse struct {
	Nonce                      string `json:"nonce"`
	TimestampMs                int64  `json:"timestampMs"`
	ApkPackageName             string `json:"apkPackageName"`
	ApkDigestSha256            string `json:"apkDigestSha256"`
	CtsProfileMatch            bool   `json:"ctsProfileMatch"`
	ApkCertificateDigestSha256 []any  `json:"apkCertificateDigestSha256"`
	BasicIntegrity             bool   `json:"basicIntegrity"`
}
```

#### type [ServerResponse](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/options.go#L268) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ServerResponse "Go to ServerResponse")

```
type ServerResponse struct {
	Status  ServerResponseStatus `json:"status"`
	Message string               `json:"errorMessage"`
}
```

#### type [ServerResponseStatus](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/options.go#L273) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#ServerResponseStatus "Go to ServerResponseStatus")

```
type ServerResponseStatus string
```

```
const (
	StatusOk     ServerResponseStatus = "ok"
	StatusFailed ServerResponseStatus = "failed"
)
```

#### type [TokenBinding](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/client.go#L38) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#TokenBinding "Go to TokenBinding")

```
type TokenBinding struct {
	Status TokenBindingStatus `json:"status"`
	ID     string             `json:"id,omitempty"`
}
```

#### type [TokenBindingStatus](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/client.go#L43) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#TokenBindingStatus "Go to TokenBindingStatus")

```
type TokenBindingStatus string
```

```
const (
	// Present indicates token binding was used when communicating with the
	// Relying Party. In this case, the id member MUST be present.
	Present TokenBindingStatus = "present"

	// Supported indicates token binding was used when communicating with the
	// negotiated when communicating with the Relying Party.
	Supported TokenBindingStatus = "supported"

	// NotSupported indicates token binding not supported
	// when communicating with the Relying Party.
	NotSupported TokenBindingStatus = "not-supported"
)
```

#### type [TopOriginVerificationMode](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/client.go#L173) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#TopOriginVerificationMode "Go to TopOriginVerificationMode")added in **v0.11.0**

```
type TopOriginVerificationMode int
```

```
const (
	// TopOriginDefaultVerificationMode represents the default verification mode for the Top Origin. At this time this
	// mode is the same as TopOriginIgnoreVerificationMode until such a time as the specification becomes stable. This
	// value is intended as a fallback value and implementers should very intentionally pick another option if they want
	// stability.
	TopOriginDefaultVerificationMode TopOriginVerificationMode = iota

	// TopOriginIgnoreVerificationMode ignores verification entirely.
	TopOriginIgnoreVerificationMode

	// TopOriginAutoVerificationMode represents the automatic verification mode for the Top Origin. In this mode the
	// If the Top Origins parameter has values it checks against this, otherwise it checks against the Origins parameter.
	TopOriginAutoVerificationMode

	// TopOriginImplicitVerificationMode represents the implicit verification mode for the Top Origin. In this mode the
	// Top Origin is verified against the allowed Origins values.
	TopOriginImplicitVerificationMode

	// TopOriginExplicitVerificationMode represents the explicit verification mode for the Top Origin. In this mode the
	// Top Origin is verified against the allowed Top Origins values.
	TopOriginExplicitVerificationMode
)
```

#### type [URLEncodedBase64](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/base64.go#L12) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#URLEncodedBase64 "Go to URLEncodedBase64")

```
type URLEncodedBase64 []byte
```

URLEncodedBase64 represents a byte slice holding URL-encoded base64 data. When fields of this type are unmarshalled from JSON, the data is base64 decoded into a byte slice.

#### func [CreateChallenge](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/challenge.go#L12) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#CreateChallenge "Go to CreateChallenge")

```
func CreateChallenge() (challenge URLEncodedBase64, err error)
```

CreateChallenge creates a new challenge that should be signed and returned by the authenticator. The spec recommends using at least 16 bytes with 100 bits of entropy. We use 32 bytes.

#### func (URLEncodedBase64) [MarshalJSON](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/base64.go#L46) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#URLEncodedBase64.MarshalJSON "Go to URLEncodedBase64.MarshalJSON")

```
func (e URLEncodedBase64) MarshalJSON() ([]byte, error)
```

MarshalJSON base64 encodes a non URL-encoded value, storing the result in the provided byte slice.

#### func (URLEncodedBase64) [String](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/base64.go#L14) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#URLEncodedBase64.String "Go to URLEncodedBase64.String")added in **v0.6.0**

```
func (e URLEncodedBase64) String() string
```

#### func (\*URLEncodedBase64) [UnmarshalJSON](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/base64.go#L20) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#URLEncodedBase64.UnmarshalJSON "Go to URLEncodedBase64.UnmarshalJSON")

```
func (e *URLEncodedBase64) UnmarshalJSON(data []byte) error
```

UnmarshalJSON base64 decodes a URL-encoded value, storing the result in the provided byte slice.

#### type [UserEntity](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/entities.go#L34) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#UserEntity "Go to UserEntity")

```
type UserEntity struct {
	CredentialEntity
	// A human-palatable name for the user account, intended only for display.
	// For example, "Alex P. Müller" or "田中 倫". The Relying Party SHOULD let
	// the user choose this, and SHOULD NOT restrict the choice more than necessary.
	DisplayName string `json:"displayName"`

	// ID is the user handle of the user account entity. To ensure secure operation,
	// authentication and authorization decisions MUST be made on the basis of this id
	// member, not the displayName nor name members. See Section 6.1 of
	// [RFC8266](https://www.w3.org/TR/webauthn/#biblio-rfc8266).
	ID any `json:"id"`
}
```

The UserEntity represents the PublicKeyCredentialUserEntity IDL and is used to supply additional user account attributes when creating a new credential.

Specification: §5.4.3 User Account Parameters for Credential Generation ([https://www.w3.org/TR/webauthn/#dictdef-publickeycredentialuserentity](https://www.w3.org/TR/webauthn/#dictdef-publickeycredentialuserentity))

#### type [UserVerificationRequirement](https://github.com/go-webauthn/webauthn/blob/v0.15.0/protocol/authenticator.go#L195) [¶](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.15.0/protocol#UserVerificationRequirement "Go to UserVerificationRequirement")

```
type UserVerificationRequirement string
```

UserVerificationRequirement is a representation of the UserVerificationRequirement IDL enum.

A WebAuthn Relying Party may require user verification for some of its operations but not for others, and may use this type to express its needs.

Specification: §5.8.6. User Verification Requirement Enumeration ([https://www.w3.org/TR/webauthn/#enum-userVerificationRequirement](https://www.w3.org/TR/webauthn/#enum-userVerificationRequirement))

```
const (
	// VerificationRequired User verification is required to create/release a credential.
	VerificationRequired UserVerificationRequirement = "required"

	// VerificationPreferred User verification is preferred to create/release a credential.
	VerificationPreferred UserVerificationRequirement = "preferred" // This is the default.

	// VerificationDiscouraged The authenticator should not verify the user for the credential.
	VerificationDiscouraged UserVerificationRequirement = "discouraged"
```
