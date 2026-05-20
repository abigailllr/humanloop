import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../main.dart';
import '../models/user.dart';
import '../services/api_service.dart';
import '../services/auth_service.dart';

class AuthState {
  final AppUser? user;
  final String? token;
  final String? refreshToken;
  final bool loading;

  const AuthState({this.user, this.token, this.refreshToken, this.loading = false});

  bool get isSignedIn => user != null;

  AuthState copyWith({AppUser? user, String? token, String? refreshToken, bool? loading}) {
    return AuthState(
      user: user ?? this.user,
      token: token ?? this.token,
      refreshToken: refreshToken ?? this.refreshToken,
      loading: loading ?? this.loading,
    );
  }
}

class AuthNotifier extends Notifier<AuthState> {
  @override
  AuthState build() {
    final prefs = ref.read(prefsProvider);
    final accessToken = prefs.getString('auth_token');
    final storedRefresh = prefs.getString('refresh_token');
    if (accessToken != null) {
      Future.microtask(() => _restoreFromToken(accessToken, storedRefresh));
      return AuthState(token: accessToken, refreshToken: storedRefresh);
    }
    return const AuthState();
  }

  Future<void> _restoreFromToken(String token, String? refreshToken) async {
    try {
      final profile = await ApiService.getProfile(token: token);
      if (profile.isNotEmpty) {
        state = AuthState(user: AppUser.fromJson(profile), token: token, refreshToken: refreshToken);
        return;
      }
    } catch (_) {}
    if (refreshToken != null) {
      final tokens = await ApiService.refreshTokens(refreshToken);
      if (tokens != null) {
        final profile = await ApiService.getProfile(token: tokens.accessToken);
        if (profile.isNotEmpty) {
          final prefs = ref.read(prefsProvider);
          await prefs.setString('auth_token', tokens.accessToken);
          await prefs.setString('refresh_token', tokens.refreshToken);
          state = AuthState(user: AppUser.fromJson(profile), token: tokens.accessToken, refreshToken: tokens.refreshToken);
          return;
        }
      }
    }
    final prefs = ref.read(prefsProvider);
    await prefs.remove('auth_token');
    await prefs.remove('refresh_token');
    state = const AuthState();
  }

  Future<void> signInWithGoogle() async {
    state = state.copyWith(loading: true);
    try {
      final result = await AuthService.signInWithGoogle();
      final prefs = ref.read(prefsProvider);
      await prefs.setString('auth_token', result.tokens.accessToken);
      await prefs.setString('refresh_token', result.tokens.refreshToken);
      state = AuthState(user: result.user, token: result.tokens.accessToken, refreshToken: result.tokens.refreshToken);
    } catch (e) {
      state = state.copyWith(loading: false);
      rethrow;
    }
  }

  Future<void> signInWithApple() async {
    state = state.copyWith(loading: true);
    try {
      final result = await AuthService.signInWithApple();
      final prefs = ref.read(prefsProvider);
      await prefs.setString('auth_token', result.tokens.accessToken);
      await prefs.setString('refresh_token', result.tokens.refreshToken);
      state = AuthState(user: result.user, token: result.tokens.accessToken, refreshToken: result.tokens.refreshToken);
    } catch (e) {
      state = state.copyWith(loading: false);
      rethrow;
    }
  }

  void devBypass() {
    const mockUser = AppUser(id: 'dev', email: 'dev@humanloop.ai', name: 'Dev User', credits: 0, submissions: 0);
    state = const AuthState(user: mockUser, token: 'dev-token');
  }

  Future<void> signOut() async {
    final token = state.token;
    final refresh = state.refreshToken;
    await AuthService.signOut();
    if (token != null && refresh != null) {
      await ApiService.revokeToken(token, refresh);
    }
    final prefs = ref.read(prefsProvider);
    await prefs.remove('auth_token');
    await prefs.remove('refresh_token');
    state = const AuthState();
  }
}

final authProvider = NotifierProvider<AuthNotifier, AuthState>(AuthNotifier.new);
