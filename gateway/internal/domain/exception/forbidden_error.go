package exception

type ForbiddenType string

const (
	ForbiddenTypeBannedUser            ForbiddenType = "banned_user"
	ForbiddenTypeUnapprovedCorporation ForbiddenType = "unapproved_corporation"
	ForbiddenTypeNoPropertyAccess      ForbiddenType = "no_property_access"
)

type ForbiddenError struct {
	Type     ForbiddenType
	Resource string
	Message  string
}

func (e ForbiddenError) Error() string {
	if e.Message != "" {
		return "Forbidden: " + e.Message
	}
	return "Forbidden: Access to " + e.Resource + " is not allowed"
}

func NewBannedUserForbiddenError() ForbiddenError {
	return ForbiddenError{
		Type:    ForbiddenTypeBannedUser,
		Message: "Your account is banned and cannot perform this operation.",
	}
}

func NewUnapprovedCorporationForbiddenError() ForbiddenError {
	return ForbiddenError{
		Type:    ForbiddenTypeUnapprovedCorporation,
		Message: "Vendor approval is required to access this resource.",
	}
}

func NewNoPropertyAccessForbiddenError(resource string) ForbiddenError {
	return ForbiddenError{
		Type:     ForbiddenTypeNoPropertyAccess,
		Resource: resource,
		Message:  "You do not have access to the property: " + resource,
	}
}
