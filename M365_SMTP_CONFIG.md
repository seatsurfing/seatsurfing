# Microsoft 365 SMTP Configuration

## Issues with M365 SMTP

The previous implementation had several compatibility issues with Microsoft 365 SMTP servers:

1. **Limited Authentication Methods**: Only supported PLAIN auth, while M365 often requires LOGIN auth
2. **Incorrect Port Configuration**: Used default port 25 instead of M365's standard port 587
3. **Missing STARTTLS Requirements**: M365 requires STARTTLS encryption
4. **Authentication Sequence Issues**: M365 is sensitive to the order of SMTP commands

## Fixes Applied

### 1. Added LOGIN Authentication Support
- Implemented `LoginAuth` struct to support LOGIN SASL mechanism
- Added automatic fallback from PLAIN to LOGIN for better M365 compatibility

### 2. Added SMTP_AUTH_METHOD Configuration
- New environment variable: `SMTP_AUTH_METHOD`
- Supported values: `PLAIN`, `LOGIN`
- Default: `PLAIN` (for backward compatibility)

### 3. Automatic M365 Detection and Optimization
- Detects M365 SMTP servers (`smtp.office365.com`, etc.)
- Automatically applies optimal settings:
  - Port 587 instead of 25
  - Forces STARTTLS
  - Prefers LOGIN authentication

### 4. Improved Error Handling
- Better logging of authentication failures
- Automatic retry with different auth methods

## Recommended M365 Configuration

```bash
MAIL_SERVICE=smtp
SMTP_HOST=smtp.office365.com
SMTP_PORT=587
SMTP_START_TLS=1
SMTP_AUTH=1
SMTP_AUTH_METHOD=LOGIN
SMTP_AUTH_USER=your-email@yourdomain.com
SMTP_AUTH_PASS=your-app-password
MAIL_SENDER_ADDRESS=your-email@yourdomain.com
```

## Security Notes

- Use App Passwords instead of regular passwords when possible
- Ensure STARTTLS is enabled (`SMTP_START_TLS=1`)
- Never set `SMTP_INSECURE_SKIP_VERIFY=1` in production

## Troubleshooting

If emails still fail to send:

1. Verify your M365 account has SMTP AUTH enabled
2. Use an App Password instead of your regular password
3. Check that your firewall allows outbound connections to port 587
4. Enable debug logging to see detailed SMTP conversation
5. Try both PLAIN and LOGIN authentication methods