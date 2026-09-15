# Security Policy

## Supported Versions

Only the latest stable release is supported. Please make sure you're upgrading regularly.

## Reporting a Vulnerability

If you find a vulnerability, please contact security@seatsurfing.io. We'll look into it shortly and fix it asap.

When reporting, please include:

* A clear description of the vulnerability and its potential impact.
* Step-by-step instructions to reproduce the issue (proof-of-concept code, requests, or scripts are welcome).
* The affected component, version, URL, or deployment (if known).
* Any relevant logs, screenshots, or output.

We ask that you:

* Give us a reasonable amount of time to investigate and remediate an issue before disclosing it publicly.
* Make a good faith effort to avoid privacy violations, destruction of data, and interruption or degradation of our services during your research.
* Only interact with accounts, data, and systems you own or for which you have explicit permission, and avoid accessing, modifying, or exfiltrating data belonging to other users.
* Not perform testing that could negatively impact Seatsurfing or its users, including (but not limited to) denial-of-service attacks, spam, social engineering or phishing of our staff or users, and physical attacks against our facilities.
* Not exploit a vulnerability beyond what is necessary to confirm and document it.

We will not pursue legal action against researchers who discover and report vulnerabilities in good faith and in accordance with this policy.

## Bug Bounty Program

We appreciate the efforts of security researchers who help us keep Seatsurfing and our users safe, and we are willing to consider compensation for verified vulnerabilities that present a genuine security risk — particularly issues that could expose sensitive customer data, lead to unauthorized access, or otherwise have a material impact on the security of our systems and users.

### How it works

1. Submit your report as described above under [Reporting a Vulnerability](#reporting-a-vulnerability).
2. Our security team will review, verify, and assess the severity and impact of the reported issue.
3. Once our assessment is complete, we will determine whether the issue qualifies for a reward and, if so, the appropriate amount, based on its severity, impact, and overall significance to Seatsurfing and our users.

### Eligibility and scope

To be eligible for consideration, a report must:

* Describe a previously unreported, verifiable vulnerability in a currently supported version of Seatsurfing's software or in infrastructure that we operate.
* Not be the result of testing prohibited under this policy (e.g. social engineering, physical attacks, denial-of-service testing, or automated scanning that generates excessive traffic).
* Not involve issues that have already been reported by another party, are already known to us, or are publicly disclosed.
* Not involve vulnerabilities in third-party services, products, or dependencies that are outside of our control (these should be reported to the respective vendor or maintainer).

Findings that are generally **not eligible**, unless a credible, demonstrated impact is shown, include (non-exhaustive):

* Missing security headers or best-practice hardening suggestions without a demonstrated exploit.
* Issues requiring physical access to a device, a compromised/rooted device, or a compromised user account.
* Reports from automated scanners without manual verification or a working proof of concept.
* Self-XSS, clickjacking on pages with no sensitive actions, or issues requiring unlikely user interaction.
* Rate limiting or brute-force issues without a demonstrated, meaningful impact.
* Vulnerabilities affecting outdated or unsupported versions.

### Rewards are entirely at our discretion

**Any decision regarding whether a bounty is awarded, as well as the amount of any compensation, remains entirely at our sole discretion.** Submitting a vulnerability report does not, by itself, create any entitlement or right to a reward. We evaluate each report individually and may decide not to offer a reward even where a report is valid, for example (without limitation) where the vulnerability has low real-world impact, is out of scope, was already known to us, or does not meet the eligibility criteria above.

We may, at our discretion, also recognize valid reports that do not qualify for a monetary reward through public acknowledgment (e.g. a hall-of-fame credit), with the reporter's consent.

### Payment of rewards

If we decide to award a reward, payment requires a valid invoice from the reporter, addressed to Seatsurfing and stating the amount agreed with our security team, before any payment can be made. As we operate in Germany, the invoice must comply with the applicable German legal and tax requirements (e.g. §14 UStG), including at minimum the reporter's full name and address, our company's name and address, the invoice date, a unique invoice number, a description of the service rendered, the net amount, the applicable VAT rate and amount (or a reference to the relevant VAT exemption, e.g. for the Kleinunternehmerregelung under §19 UStG, if applicable), and the reporter's tax identification number or VAT ID where required. It is the reporter's responsibility to ensure the invoice meets these requirements and to bear any tax obligations arising from the reward. We reserve the right to withhold payment until a compliant invoice is provided, and we may request additional documentation (e.g. identity verification) as required for payment processing or legal/regulatory compliance.

### Confidentiality

Please keep any vulnerability details confidential until we have confirmed that the issue has been resolved, and coordinate the timing of any public disclosure with us in advance. We may update this policy from time to time; the version published in this repository at the time of your report governs that report.
