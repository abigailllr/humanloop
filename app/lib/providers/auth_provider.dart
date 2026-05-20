import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../main.dart';
import '../models/user.dart';
import '../services/api_service.dart';
import '../services/auth_service.dart';

class AuthState {
  final AppUser? user;
  final String? token;
  final bool loading;

  const AuthState({this.user, this.token, this.loading = false});

  bool get isSignedIn => user != null;

  AuthState copyWith({AppUser? user, String? token, bool? loading}) {
    return AuthState(
      user: user ?? this.user,
      token: token ?? this.token,
      loading: loading ?? this.loading,
    );
  }
}

class AuthNotifier extends Notifier<AuthState> {
  @override
  AuthState build() {
    final stored = ref.read(prefsProvider).getString('auth_token');
    if (stored != null) {
      Future.microtask(() => _restoreFromToken(stored));
      return AuthState(token: stored);
    }
    return const AuthState();
  }

  Future<void> _restoreFromToken(String token) async {
    try {
      final profile = await ApiService.getProfile(token: token);
      if (profile.isNotEmpty) {
        state = AuthState(user: AppUser.fromJson(profile), token: token);
        return;
      }
    } catch (_) {}
    ref.read(prefsProvider).remove('auth_token');
    state = const AuthState();
  }

  Future<void> signInWithGoogle() async {
    state = state.copyWith(loading: true);
    try {
      final result = await AuthService.signInWithGoogle();
      ref.read(prefsProvider).setString('auth_token', result.token);
      state = AuthState(user: result.user, token: result.token);
    } catch (e) {
      state = state.copyWith(loading: false);
      rethrow;
    }
  }

  Future<void> signInWithApple() async {
    state = state.copyWith(loading: true);
    try {
      final result = await AuthService.signInWithApple();
      ref.read(prefsProvider).setString('auth_token', result.token);
      state = AuthState(user: result.user, token: result.token);
    } catch (e) {
      state = state.copyWith(loading: false);
      rethrow;
    }
  }

  void devBypass() {
    const mockUser = AppUser(id: 'dev', email: 'dev@humanloop.ai', name: 'Dev User', credits: 0, submissions: 0);
    state = AuthState(user: mockUser, token: 'dev-token');
  }

  Future<void> signOut() async {
    await AuthService.signOut();
    ref.read(prefsProvider).remove('auth_token');
    state = const AuthState();
  }
}

final authProvider = NotifierProvider<AuthNotifier, AuthState>(AuthNotifier.new);
