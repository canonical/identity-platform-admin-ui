package authorization

import (
	"context"
)

func (a *Authorizer) createTuple(ctx context.Context, resourceID string) error {
	user, ok := ctx.Value(USER_CTX).(*User)
	if !ok {
		// Should we panic?
		return nil
	}
	go func() {
		err := a.client.WriteTuple(context.Background(), "user:"+user.ID, CAN_VIEW, resourceID)
		if err != nil {
			a.logger.Errorf("Failed to create authorization tuple: %s", err)
		}
	}()
	return nil
}

func (a *Authorizer) deleteTuple(ctx context.Context, resourceID string) error {
	user, ok := ctx.Value(USER_CTX).(*User)
	if !ok {
		// Should we panic?
		return nil
	}
	go func() {
		err := a.client.DeleteTuple(context.Background(), "user:"+user.ID, CAN_VIEW, resourceID)
		if err != nil {
			a.logger.Errorf("Failed to create authorization tuple: %s", err)
		}
	}()
	return nil
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
