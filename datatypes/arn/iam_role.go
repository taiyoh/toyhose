package arn

import (
	"errors"
	"fmt"
	"strings"
)

// arn:aws:iam::account-id:role/role-name

// IAMRole provides AWS ARN for IAMRole
type IAMRole struct {
	accountID string
	roleName  string
}

// Code returns IAMRole ARN as string
func (r IAMRole) Code() string {
	return fmt.Sprintf("arn:aws:iam::%s:role/%s", r.accountID, r.roleName)
}

// RestoreIAMRoleFromRaw returns IAMRole ARN. also error returns if it is invalid
func RestoreIAMRoleFromRaw(raw string) (IAMRole, error) {
	prefix := "arn:aws:iam::"
	if !strings.HasPrefix(raw, prefix) {
		return IAMRole{}, errors.New("invalid prefix")
	}
	cut := strings.Replace(raw, prefix, "", 1)
	elems := strings.Split(cut, ":")
	if len(elems) != 2 {
		return IAMRole{}, errors.New("no role assigned")
	}
	accountID := elems[0]
	roleResource := elems[1]
	rolePrefix := "role/"
	if !strings.HasPrefix(roleResource, rolePrefix) {
		return IAMRole{}, errors.New("invalid role assign description")
	}
	roleName := strings.Replace(roleResource, rolePrefix, "", 1)
	return IAMRole{accountID, roleName}, nil
}
