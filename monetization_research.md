# Monetization Strategy Research for ATS-Tex Lab

Based on the existing layout of your application and database architecture, you hold several primary value propositions. The most critical aspect is the episodic nature of job hunting—as you mentioned, users don't need this tool every single month.

## 1. Existing Monetizable Features
1. **AI-Powered Features** (High API cost for you, highly valuable for users):
   - **CV Reviews** (Scoring, recommendations, strengths/weaknesses)
   - **Cover Letter Generation** 
   - **ATS Simulator** (Matching CVs with specific Job Descriptions)
2. **Core Utility Features** (Low cost for you):
   - Creating multiple CV Profiles
   - Job Application Tracking (Kanban board)
   - Public CV Profiles

## 2. Monetization Mechanisms Compared

### A. Token/Credit-Based System (Pay-As-You-Go)
In this model, users purchase a pool of "AI Credits" (e.g., 500 credits for $5). Each time they use an AI feature, it deducts a specific amount of credits based on its cost complexity.

> [!TIP]
> **Why this fits your app perfectly:** Your database already tracks `ai_tokens_used` in the `users`, `cv_reviews`, and `cover_letters` tables! The infrastructure is almost entirely there.

- **Pros:** 
  - **Ideal for episodic use**: Users don't feel "cheated" by a monthly subscription they only use once a week. They pay only for what they use.
  - Very easy to balance. If OpenAI/Anthropic changes prices, you simply tweak how many credits a feature costs.
- **Cons:** 
  - Revenue is less predictable for you compared to MRR (Monthly Recurring Revenue).
  - Users can experience "token anxiety," hesitating to generate cover letters to save credits.

### B. Fixed Monthly/Yearly Tiered Subscriptions
Users pay a recurring fee (e.g., Free, $9/mo Pro tier) to get unlimited or high-limit access.

- **Pros:** 
  - Predictable, recurring revenue.
  - Easier to market (e.g., "Get Unlimited AI Cover Letters!").
- **Cons:** 
  - **The "Cancel Problem":** Users looking for jobs usually find one within 1-3 months and will immediately cancel. The churn rate will be naturally extremely high.

### C. Time-Limited Passes (Non-Renewing)
A fantastic hybrid option for career tools. Instead of auto-renewing subscriptions, users buy a "pass" for a fixed timeframe.
- E.g., **"7-Day Interview Sprint Pass ($5)"** or **"30-Day Job Hunter Pass ($12)."** 
- Once it expires, they drop back to the Free tier. This fully eliminates user subscription anxiety while providing you upfront cash.

## 3. Recommended Approach

Given your preference ("good cuz we not used it always"), the most ethical and user-friendly approach is a **Token Package System combined with a freemium baseline.**

### The Recommended Setup

**1. Free Tier (To acquire users):**
- 1 CV Profile
- Up to 10 Jobs tracked in the Kanban Board
- **50 Free Starter Credits** (Enough to preview the AI capabilities: e.g., 1 CV Review, 1 Cover Letter, 1 ATS Simulation)

**2. Token Packages (One-Time Purchases):**
- **Starter Pack ($4.99):** 250 AI Credits
- **Pro Pack ($9.99):** 700 AI Credits

**3. Feature Credit Pricing Strategy:**
- **ATS Simulation:** 20 credits (Requires analyzing JD + CV thoroughly)
- **Cover Letter Gen:** 15 credits
- **CV Review:** 15 credits
- **Unlock Additional CV Profiles:** 50 credits (Permanent unlock to give tokens utility outside just AI generation)
- **Job Kanban Board:** Keep free and unlimited to encourage users to keep opening your app daily.

## 4. Alternative: Feature-Quota Subscription
If you explicitly prefer a monthly model where they get "N total AI reviews for subscription A":

- **Free Tier:** 1 CV Profile, 1 AI CV Review/mo, 1 ATS Simulation/mo.
- **Basic Tier ($5/mo):** 3 CV Profiles, 10 AI CV Reviews/mo, 10 ATS Simulations/mo, 10 Cover Letters/mo.
- **Pro Tier ($12/mo):** Unlimited Profiles, 50 AI CV Reviews/mo, 50 ATS Simulations/mo, 50 Cover Letters/mo.

## Summary & Next Steps
From a technical standpoint, going with a **Token/Credit System** is incredibly fast to implement for you, because you already have `ai_tokens_used` across your tables. We would only need to:
1. Add a `credits_balance` column to the `users` table.
2. Build a middleware to check `credits_balance >= required_credits` before executing ai-handlers.
3. Integrate a payment gateway (like Stripe, Midtrans, or Xendit) to refill this balance.

Which approach aligns best with your vision for the product?
