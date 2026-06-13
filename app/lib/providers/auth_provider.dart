import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/user.dart';
import '../services/api_service.dart';
import '../services/auth_service.dart';
import '../services/token_store.dart';

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
    Future.microtask(_init);
    return const AuthState(loading: true);
  }

  Future<void> _init() async {
    final accessToken = await TokenStore.readAccess();
    final storedRefresh = await TokenStore.readRefresh();
    if (accessToken == null) {
      state = const AuthState();
      return;
    }
    await _restoreFromToken(accessToken, storedRefresh);
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
          await TokenStore.save(tokens.accessToken, tokens.refreshToken);
          state = AuthState(user: AppUser.fromJson(profile), token: tokens.accessToken, refreshToken: tokens.refreshToken);
          return;
        }
      }
    }
    await TokenStore.clear();
    state = const AuthState();
  }

  Future<void> signInWithGoogle() async {
    state = state.copyWith(loading: true);
    try {
      final result = await AuthService.signInWithGoogle();
      await TokenStore.save(result.tokens.accessToken, result.tokens.refreshToken);
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
      await TokenStore.save(result.tokens.accessToken, result.tokens.refreshToken);
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
    await TokenStore.clear();
    state = const AuthState();
  }
}

final authProvider = NotifierProvider<AuthNotifier, AuthState>(AuthNotifier.new);
