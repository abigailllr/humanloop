import 'package:google_sign_in/google_sign_in.dart';
import 'package:sign_in_with_apple/sign_in_with_apple.dart';
import '../models/user.dart';
import 'api_service.dart';

class AuthService {
  static final _googleSignIn = GoogleSignIn(scopes: ['email', 'profile']);

  static Future<({AppUser user, AuthTokens tokens})> signInWithGoogle() async {
    final account = await _googleSignIn.signIn();
    if (account == null) throw Exception('sign-in cancelled');
    final auth = await account.authentication;
    final idToken = auth.idToken;
    if (idToken == null) throw Exception('no id token');
    final tokens = await ApiService.authGoogle(idToken);
    final profile = await ApiService.getProfile(token: tokens.accessToken);
    final user = AppUser.fromJson(profile.isNotEmpty ? profile : {'id': 'google:${account.id}', 'name': account.displayName ?? '', 'email': account.email});
    return (user: user, tokens: tokens);
  }

  static Future<({AppUser user, AuthTokens tokens})> signInWithApple() async {
    final credential = await SignInWithApple.getAppleIDCredential(
      scopes: [AppleIDAuthorizationScopes.email, AppleIDAuthorizationScopes.fullName],
    );
    final name = [credential.givenName, credential.familyName].where((s) => s != null && s.isNotEmpty).join(' ');
    final tokens = await ApiService.authApple(
      identityToken: credential.identityToken ?? '',
      userId: credential.userIdentifier ?? '',
      email: credential.email ?? '',
      name: name,
    );
    final profile = await ApiService.getProfile(token: tokens.accessToken);
    final user = AppUser.fromJson(profile.isNotEmpty ? profile : {'id': 'apple:${credential.userIdentifier}', 'name': name, 'email': credential.email ?? ''});
    return (user: user, tokens: tokens);
  }

  static Future<void> signOut() async {
    await _googleSignIn.signOut();
  }
}
