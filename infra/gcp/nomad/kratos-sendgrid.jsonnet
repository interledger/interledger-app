local getFirstName(ctx) = 
  if std.objectHas(ctx, "template_data") && std.objectHas(ctx.template_data, "identity") && std.objectHas(ctx.template_data.identity, "traits") then ctx.template_data.identity.traits.firstName else "";

local generateSubject(ctx) = 
  if "template_type" in ctx && ctx.template_type == 'recovery_invalid' then "Recover your account."
  else if "template_type" in ctx && ctx.template_type == 'recovery_valid' then "Recover your account."
  else "Fynbos";

local invalidRecoveryAttempt(ctx) = {
  data: [
    {
      heading: "Invalid account recovery attempt",
      paragraph: "You (or someone else) entered this email address when trying to recover access to a Fynbos account."
    },
    {
      paragraph: "However, this email address is not on our database of registered users and therefore the attempt has failed."
    },
    {
      paragraph: "If this was you, please check if you signed up using a different email address."
    },
    {
      paragraph: "If this was not you, you can safely ignore this email."
    }
  ],
  subject: generateSubject(ctx)
};

local validRecoveryAttempt(ctx) = {
  local firstName = getFirstName(ctx),
  data: [
    {
      paragraph: if firstName == "" then "Hello," else "Hello " + firstName + ","
    },
    {
      heading: "Password reset request",
      paragraph: "Please click the button below to reset your password."
    },
    {
      paragraph: "If you did not request to reset your password, you can safely ignore this email."
    }
  ],
  cta: {
    text: "Reset password",
    url: if std.objectHas(ctx, "template_data") && std.objectHas(ctx.template_data, "recovery_url") then ctx.template_data.recovery_url else ""
  },
  subject: generateSubject(ctx)
};

local generateTemplateData(ctx) = 
  local templateType = if std.objectHas(ctx, "template_type") then ctx.template_type else "";
  if templateType == 'recovery_invalid' then invalidRecoveryAttempt(ctx) 
  else if templateType == 'recovery_valid' then validRecoveryAttempt(ctx)
  else {};

function(ctx) {
  from: {email: 'hello@fynbos.app', name: 'Fynbos'},
  subject: generateSubject(ctx),
  personalizations: [
    {
      to: [{email: if "template_data" in ctx && "to" in ctx.template_data then ctx.template_data.to else ""}],
      dynamic_template_data: generateTemplateData(ctx)
    }
  ],
  template_id: "d-d1d84d89553a43f89d6c60e2497b24c3"
}
