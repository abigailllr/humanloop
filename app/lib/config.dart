class AppConfig {
  static const apiBaseUrl = String.fromEnvironment(
    'API_URL',
    defaultValue: 'https://api.humanloop.app',
  );

  static const privacyPolicyUrl = String.fromEnvironment(
    'PRIVACY_URL',
    defaultValue: 'https://humanloop.app/privacy',
  );

  static const termsUrl = String.fromEnvironment(
    'TERMS_URL',
    defaultValue: 'https://humanloop.app/terms',
  );

  static const supportEmail = 'support@humanloop.app';
}
