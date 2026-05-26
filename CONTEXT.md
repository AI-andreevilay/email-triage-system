# Email Triage System

Email Triage System classifies email messages for a user and applies labels that help the user review or process their inbox.

## Language

**User Rule**:
A classification rule that is applicable to users. A User Rule can be global for every user or specific to one user.
_Avoid_: Custom rule when the distinction between global and user-specific matters

**Global Rule**:
A User Rule that applies to every user and provides the system-wide default classification behavior.
_Avoid_: Built-in rule

**User-Specific Rule**:
A User Rule that applies to exactly one user. A User-Specific Rule expresses that user's explicit intent and takes precedence over Global Rules.
_Avoid_: Personal rule

**Mailbox Owner**:
The person whose inbox is being triaged. In the current product scope, the Mailbox Owner and the administrator can be the same person.
_Avoid_: Customer

**Single-Owner Scope**:
The current product scope where the system is operated for one Mailbox Owner. The language still distinguishes Global Rules from User-Specific Rules so the model can grow beyond one owner later.
_Avoid_: Single-user SaaS

**Built-in Rule**:
A classification rule embedded in application code rather than managed as a User Rule.
_Avoid_: Default rule

## Example Dialogue

Developer: Should this sender rule be a Global Rule?

Domain expert: Yes, it should classify the same sender for every user.

Developer: Should we keep the current classification rules as Built-in Rules?

Domain expert: No. Move them into Global Rules so an admin can manage the default behavior.

Developer: Is the admin console separate from the Mailbox Owner experience?

Domain expert: Not in the current scope. The Mailbox Owner can use the admin console directly.

Developer: Are User-Specific Rules still meaningful in Single-Owner Scope?

Domain expert: Yes. They represent the Mailbox Owner's overrides above Global Rules.
