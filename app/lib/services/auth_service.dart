import '../models/user.dart';

/// Mock auth service. Replace sign-in methods with real OAuth + backend JWT
/// when the backend is ready.
class AuthService {
  static AppUser? _currentUser;

  static AppUser? get currentUser => _currentUser;
  static bool get isSignedIn => _currentUser != null;

  /// Mock Google Sign-In — navigates to main app without real OAuth.
  static Future<AppUser> signInWithGoogle() async {
    await Future.delayed(const Duration(milliseconds: 600));
    _currentUser = const AppUser(
      id: 'google-mock-001',
      name: 'Alex Johnson',
      email: 'alex@gmail.com',
      credits: 4,
      submissions: 3,
      rank: 42,
    );
    return _currentUser!;
  }

  /// Mock Apple Sign-In — navigates to main app without real OAuth.
  static Future<AppUser> signInWithApple() async {
    await Future.delayed(const Duration(milliseconds: 600));
    _currentUser = const AppUser(
      id: 'apple-mock-001',
      name: 'Alex Johnson',
      email: 'alex@privaterelay.appleid.com',
      credits: 4,
      submissions: 3,
      rank: 42,
    );
    return _currentUser!;
  }

  static Future<void> signOut() async {
    _currentUser = null;
  }
}
