package authorization

import (
	"context"
	"fmt"
)

var ErrNoUserInContext = fmt.Errorf("user not authenticated or requests context broken, cannot perform operation")

func (a *Authorizer) createTuple(ctx context.Context, resourceID string) error {
	user, ok := ctx.Value(USER_CTX).(*User)
	if !ok {
		return ErrNoUserInContext
	}

	_, err := a.wpool.Submit(
		func() any {
			if err := a.client.WriteTuple(context.Background(), "user:"+user.ID, CAN_VIEW, resourceID); err != nil {
				a.logger.Error("Async write failed: ", err.Error())
				return err
			}
			return nil
		},
		nil,
		nil,
	)
	if err != nil {
		a.logger.Errorf("Failed to submit job to worker pool: %s", err)
	}
	return err
}

func (a *Authorizer) deleteTuple(ctx context.Context, resourceID string) error {
	user, ok := ctx.Value(USER_CTX).(*User)
	if !ok {
		return ErrNoUserInContext
	}

	_, err := a.wpool.Submit(
		func() any {
			if err := a.client.DeleteTuple(context.Background(), "user:"+user.ID, CAN_VIEW, resourceID); err != nil {
				a.logger.Error("Async delete failed: ", err.Error())
				return err
			}
			return nil
		},
		nil,
		nil,
	)
	if err != nil {
		a.logger.Errorf("Failed to submit job to worker pool: %s", err)
	}
	return err
}

func (a *Authorizer) getResource(resourceID, resourceType string) string {
	return resourceType + ":" + resourceID
}

func (a *Authorizer) SetCreateClientEntitlements(ctx context.Context, clientID string) error {
	return a.createTuple(ctx, a.getResource(clientID, CLIENT_TYPE))
}

func (a *Authorizer) SetDeleteClientEntitlements(ctx context.Context, clientID string) error {
	return a.deleteTuple(ctx, a.getResource(clientID, CLIENT_TYPE))
}

func (a *Authorizer) SetCreateIdentityEntitlements(ctx context.Context, IdentityID string) error {
	return a.createTuple(ctx, a.getResource(IdentityID, IDENTITY_TYPE))
}

func (a *Authorizer) SetDeleteIdentityEntitlements(ctx context.Context, IdentityID string) error {
	return a.deleteTuple(ctx, a.getResource(IdentityID, IDENTITY_TYPE))
}

func (a *Authorizer) SetCreateProviderEntitlements(ctx context.Context, providerID string) error {
	return a.createTuple(ctx, a.getResource(providerID, PROVIDER_TYPE))
}

func (a *Authorizer) SetDeleteProviderEntitlements(ctx context.Context, providerID string) error {
	return a.deleteTuple(ctx, a.getResource(providerID, PROVIDER_TYPE))
}

func (a *Authorizer) SetCreateRuleEntitlements(ctx context.Context, ruleID string) error {
	return a.createTuple(ctx, a.getResource(ruleID, RULE_TYPE))
}

func (a *Authorizer) SetDeleteRuleEntitlements(ctx context.Context, ruleID string) error {
	return a.deleteTuple(ctx, a.getResource(ruleID, RULE_TYPE))
}

func (a *Authorizer) SetCreateSchemaEntitlements(ctx context.Context, schemeID string) error {
	return a.createTuple(ctx, a.getResource(schemeID, SCHEME_TYPE))
}

func (a *Authorizer) SetDeleteSchemaEntitlements(ctx context.Context, schemeID string) error {
	return a.deleteTuple(ctx, a.getResource(schemeID, SCHEME_TYPE))
}
