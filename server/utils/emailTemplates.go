package utils

import (
	"fmt"

	"github.com/leal-hospital/server/domain"
)

// GenerateOTPEmail generates OTP email with subject and body
func GenerateOTPEmail(name, otp string) domain.EmailContent {
	return domain.EmailContent{
		Subject: "Your OTP for Lael Hospital",
		Body: fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>OTP Verification - Lael Hospital</title>
</head>
<body style="margin: 0; padding: 0; font-family: Arial, sans-serif; background-color: #f4f4f4;">
    <table role="presentation" style="width: 100%%; border-collapse: collapse;">
        <tr>
            <td align="center" style="padding: 40px 0;">
                <table role="presentation" style="width: 600px; border-collapse: collapse; background-color: #ffffff; box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);">
                    <!-- Header with purple gradient -->
                    <tr>
                        <td style="background: linear-gradient(135deg, #5B2C91 0%%, #7B3FA6 100%%); padding: 40px 30px; text-align: center;">
                            <h1 style="margin: 0; color: #ffffff; font-size: 28px; font-weight: bold;">Lael Hospital</h1>
                            <p style="margin: 10px 0 0 0; color: #ffffff; font-size: 16px;">Healthcare Management System</p>
                        </td>
                    </tr>

                    <!-- Content -->
                    <tr>
                        <td style="padding: 40px 30px;">
                            <h2 style="margin: 0 0 20px 0; color: #333333; font-size: 24px;">Hello, %s!</h2>
                            <p style="margin: 0 0 20px 0; color: #666666; font-size: 16px; line-height: 1.5;">
                                You have requested a One-Time Password (OTP) to verify your identity. Please use the code below to complete your verification:
                            </p>

                            <!-- OTP Box -->
                            <table role="presentation" style="width: 100%%; border-collapse: collapse; margin: 30px 0;">
                                <tr>
                                    <td align="center">
                                        <div style="background-color: #f8f5fc; border: 2px dashed #7B3FA6; border-radius: 8px; padding: 20px; display: inline-block;">
                                            <p style="margin: 0 0 10px 0; color: #5B2C91; font-size: 14px; font-weight: bold; text-transform: uppercase; letter-spacing: 1px;">Your OTP Code</p>
                                            <p style="margin: 0; color: #5B2C91; font-size: 36px; font-weight: bold; letter-spacing: 8px; font-family: 'Courier New', monospace;">%s</p>
                                        </div>
                                    </td>
                                </tr>
                            </table>

                            <!-- Warning Box -->
                            <table role="presentation" style="width: 100%%; border-collapse: collapse; margin: 20px 0;">
                                <tr>
                                    <td style="background-color: #fff3cd; border-left: 4px solid #ffc107; padding: 15px; border-radius: 4px;">
                                        <p style="margin: 0; color: #856404; font-size: 14px; line-height: 1.5;">
                                            <strong>⚠️ Important:</strong> This OTP is valid for only <strong>5 minutes</strong>. Do not share this code with anyone.
                                        </p>
                                    </td>
                                </tr>
                            </table>

                            <p style="margin: 20px 0 0 0; color: #666666; font-size: 14px; line-height: 1.5;">
                                If you did not request this OTP, please ignore this email or contact our support team immediately.
                            </p>
                        </td>
                    </tr>

                    <!-- Footer -->
                    <tr>
                        <td style="background-color: #f8f5fc; padding: 30px; text-align: center; border-top: 1px solid #e0e0e0;">
                            <p style="margin: 0 0 10px 0; color: #666666; font-size: 14px;">
                                Thank you for choosing Lael Hospital
                            </p>
                            <p style="margin: 0; color: #999999; font-size: 12px;">
                                This is an automated message, please do not reply to this email.
                            </p>
                            <p style="margin: 10px 0 0 0; color: #7B3FA6; font-size: 12px;">
                                © 2026 Lael Hospital. All rights reserved.
                            </p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>
`, name, otp),
	}
}

// GenerateWelcomeEmail generates HTML email for welcome message
func GenerateWelcomeEmail(name string) domain.EmailContent {
	return domain.EmailContent{
		Subject: "Welcome to Lael Hospital",
		Body: fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Welcome to Lael Hospital</title>
</head>
<body style="margin: 0; padding: 0; font-family: Arial, sans-serif; background-color: #f4f4f4;">
    <table role="presentation" style="width: 100%%; border-collapse: collapse;">
        <tr>
            <td align="center" style="padding: 40px 0;">
                <table role="presentation" style="width: 600px; border-collapse: collapse; background-color: #ffffff; box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);">
                    <!-- Header with purple gradient -->
                    <tr>
                        <td style="background: linear-gradient(135deg, #5B2C91 0%%, #7B3FA6 100%%); padding: 40px 30px; text-align: center;">
                            <h1 style="margin: 0; color: #ffffff; font-size: 28px; font-weight: bold;">Welcome to Lael Hospital</h1>
                            <p style="margin: 10px 0 0 0; color: #ffffff; font-size: 16px;">Healthcare Management System</p>
                        </td>
                    </tr>

                    <!-- Content -->
                    <tr>
                        <td style="padding: 40px 30px;">
                            <h2 style="margin: 0 0 20px 0; color: #333333; font-size: 24px;">Hello, %s!</h2>
                            <p style="margin: 0 0 20px 0; color: #666666; font-size: 16px; line-height: 1.5;">
                                We're excited to have you join the Lael Hospital family! Your account has been successfully created and verified.
                            </p>

                            <p style="margin: 0 0 20px 0; color: #666666; font-size: 16px; line-height: 1.5;">
                                With your new account, you can:
                            </p>

                            <ul style="margin: 0 0 20px 0; padding-left: 20px; color: #666666; font-size: 16px; line-height: 1.8;">
                                <li>Access your medical records securely</li>
                                <li>Schedule and manage appointments</li>
                                <li>View test results and prescriptions</li>
                                <li>Communicate with healthcare providers</li>
                                <li>Manage your health profile</li>
                            </ul>

                            <!-- CTA Button -->
                            <table role="presentation" style="width: 100%%; border-collapse: collapse; margin: 30px 0;">
                                <tr>
                                    <td align="center">
                                        <a href="#" style="display: inline-block; background: linear-gradient(135deg, #5B2C91 0%%, #7B3FA6 100%%); color: #ffffff; text-decoration: none; padding: 15px 40px; border-radius: 5px; font-size: 16px; font-weight: bold;">
                                            Get Started
                                        </a>
                                    </td>
                                </tr>
                            </table>

                            <p style="margin: 20px 0 0 0; color: #666666; font-size: 14px; line-height: 1.5;">
                                If you have any questions or need assistance, our support team is here to help you 24/7.
                            </p>
                        </td>
                    </tr>

                    <!-- Footer -->
                    <tr>
                        <td style="background-color: #f8f5fc; padding: 30px; text-align: center; border-top: 1px solid #e0e0e0;">
                            <p style="margin: 0 0 10px 0; color: #666666; font-size: 14px;">
                                Thank you for choosing Lael Hospital
                            </p>
                            <p style="margin: 0; color: #999999; font-size: 12px;">
                                This is an automated message, please do not reply to this email.
                            </p>
                            <p style="margin: 10px 0 0 0; color: #7B3FA6; font-size: 12px;">
                                © 2026 Lael Hospital. All rights reserved.
                            </p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>
`, name),
	}
}

// GeneratePasswordResetConfirmation generates HTML email for password reset confirmation
func GeneratePasswordResetConfirmation(name string) domain.EmailContent {
	return domain.EmailContent{
		Subject: "Password Reset Successful",
		Body: fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Password Reset Confirmation - Lael Hospital</title>
</head>
<body style="margin: 0; padding: 0; font-family: Arial, sans-serif; background-color: #f4f4f4;">
    <table role="presentation" style="width: 100%%; border-collapse: collapse;">
        <tr>
            <td align="center" style="padding: 40px 0;">
                <table role="presentation" style="width: 600px; border-collapse: collapse; background-color: #ffffff; box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);">
                    <!-- Header with purple gradient -->
                    <tr>
                        <td style="background: linear-gradient(135deg, #5B2C91 0%%, #7B3FA6 100%%); padding: 40px 30px; text-align: center;">
                            <h1 style="margin: 0; color: #ffffff; font-size: 28px; font-weight: bold;">Lael Hospital</h1>
                            <p style="margin: 10px 0 0 0; color: #ffffff; font-size: 16px;">Healthcare Management System</p>
                        </td>
                    </tr>

                    <!-- Content -->
                    <tr>
                        <td style="padding: 40px 30px;">
                            <h2 style="margin: 0 0 20px 0; color: #333333; font-size: 24px;">Hello, %s!</h2>
                            <p style="margin: 0 0 20px 0; color: #666666; font-size: 16px; line-height: 1.5;">
                                This email confirms that your password has been successfully reset.
                            </p>

                            <!-- Success Box -->
                            <table role="presentation" style="width: 100%%; border-collapse: collapse; margin: 20px 0;">
                                <tr>
                                    <td style="background-color: #d4edda; border-left: 4px solid #28a745; padding: 15px; border-radius: 4px;">
                                        <p style="margin: 0; color: #155724; font-size: 14px; line-height: 1.5;">
                                            <strong>✓ Success:</strong> Your password has been changed and is now active.
                                        </p>
                                    </td>
                                </tr>
                            </table>

                            <p style="margin: 20px 0; color: #666666; font-size: 16px; line-height: 1.5;">
                                You can now log in to your account using your new password. Please keep your password secure and do not share it with anyone.
                            </p>

                            <!-- Security Tips -->
                            <table role="presentation" style="width: 100%%; border-collapse: collapse; margin: 20px 0;">
                                <tr>
                                    <td style="background-color: #f8f5fc; padding: 20px; border-radius: 4px; border: 1px solid #7B3FA6;">
                                        <p style="margin: 0 0 10px 0; color: #5B2C91; font-size: 16px; font-weight: bold;">
                                            Security Tips:
                                        </p>
                                        <ul style="margin: 0; padding-left: 20px; color: #666666; font-size: 14px; line-height: 1.8;">
                                            <li>Use a strong, unique password</li>
                                            <li>Never share your password with anyone</li>
                                            <li>Enable two-factor authentication if available</li>
                                            <li>Change your password regularly</li>
                                        </ul>
                                    </td>
                                </tr>
                            </table>

                            <!-- Warning Box -->
                            <table role="presentation" style="width: 100%%; border-collapse: collapse; margin: 20px 0;">
                                <tr>
                                    <td style="background-color: #fff3cd; border-left: 4px solid #ffc107; padding: 15px; border-radius: 4px;">
                                        <p style="margin: 0; color: #856404; font-size: 14px; line-height: 1.5;">
                                            <strong>⚠️ Important:</strong> If you did not make this change, please contact our support team immediately to secure your account.
                                        </p>
                                    </td>
                                </tr>
                            </table>

                            <p style="margin: 20px 0 0 0; color: #666666; font-size: 14px; line-height: 1.5;">
                                If you have any concerns about your account security, please don't hesitate to reach out to our support team.
                            </p>
                        </td>
                    </tr>

                    <!-- Footer -->
                    <tr>
                        <td style="background-color: #f8f5fc; padding: 30px; text-align: center; border-top: 1px solid #e0e0e0;">
                            <p style="margin: 0 0 10px 0; color: #666666; font-size: 14px;">
                                Thank you for choosing Lael Hospital
                            </p>
                            <p style="margin: 0; color: #999999; font-size: 12px;">
                                This is an automated message, please do not reply to this email.
                            </p>
                            <p style="margin: 10px 0 0 0; color: #7B3FA6; font-size: 12px;">
                                © 2026 Lael Hospital. All rights reserved.
                            </p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>
`, name),
	}
}
