package main

import (
	"fmt"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateUser(ctx *pulumi.Context, name string, groups pulumi.StringArray, pgpKey string, provider *aws.Provider) error {

	user, err := iam.NewUser(ctx, fmt.Sprintf("%s-user", name), &iam.UserArgs{
		Name: pulumi.String(name),
		Path: pulumi.String("/"),
	}, pulumi.Provider(provider))
	if err != nil {
		return err
	}
	_, err = iam.NewUserGroupMembership(ctx, fmt.Sprintf("%s-groups", name), &iam.UserGroupMembershipArgs{
		User:   user.Name,
		Groups: groups,
	}, pulumi.Provider(provider))
	if err != nil {
		return err
	}
	loginProfile, err := iam.NewUserLoginProfile(ctx, fmt.Sprintf("%s-login-profile", name), &iam.UserLoginProfileArgs{
		User:                  user.Name,
		PasswordResetRequired: pulumi.Bool(true),
		PgpKey:                pulumi.String(pgpKey),
	}, pulumi.Provider(provider))
	if err != nil {
		return err
	}

	ctx.Export(fmt.Sprintf("%sLoginProfile", name), pulumi.Map{
		"user":     loginProfile.User,
		"pgpKey":   loginProfile.PgpKey,
		"password": loginProfile.EncryptedPassword,
	})
	return nil
}
