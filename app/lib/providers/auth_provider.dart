import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/user.dart';
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
  AuthState build() => const AuthState();

  Future<void> signInWithGoogle() async {
    state = state.copyWith(loading: true);
    final user = await AuthService.signInWithGoogle();
    state = AuthState(user: user, token: 'mock-jwt-google');
  }

  Future<void> signInWithApple() async {
    state = state.copyWith(loading: true);
    final user = await AuthService.signInWithApple();
    state = AuthState(user: user, token: 'mock-jwt-apple');
  }

  Future<void> signOut() async {
    await AuthService.signOut();
    state = const AuthState();
  }
}

final authProvider = NotifierProvider<AuthNotifier, AuthState>(AuthNotifier.new);
