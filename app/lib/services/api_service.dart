import 'dart:async';
import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/challenge.dart';
import '../models/submission.dart';
import '../models/user.dart';

class AuthTokens {
  final String accessToken;
  final String refreshToken;
  const AuthTokens({required this.accessToken, required this.refreshToken});
}

class ApiService {
  static const _base = String.fromEnvironment('API_URL', defaultValue: 'http://localhost:8080');
  static String get baseUrl => _base;
  static void Function()? onUnauthorized;

  static const _fallbackChallenges = [
    Challenge(id: 'c1', title: 'Pick & Place', description: 'Pick up any object from a table and place it into a box.', submissions: 0),
    Challenge(id: 'c2', title: 'Fold It', description: 'Fold a piece of cloth or paper in half.', submissions: 0),
    Challenge(id: 'c3', title: 'Sort & Stack', description: 'Sort 5 objects by size from smallest to largest.', submissions: 0, difficulty: 'Medium'),
  ];

  static Future<AuthTokens> authGoogle(String idToken) async {
    final res = await http.post(
      Uri.parse('$_base/v1/auth/google'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({'id_token': idToken}),
    );
    if (res.statusCode != 200) throw Exception('google auth failed');
    final j = jsonDecode(res.body);
    return AuthTokens(accessToken: j['access_token'] as String, refreshToken: j['refresh_token'] as String);
  }

  static Future<AuthTokens> authApple({
    required String identityToken,
    required String userId,
    String email = '',
    String name = '',
  }) async {
    final res = await http.post(
      Uri.parse('$_base/v1/auth/apple'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({'identity_token': identityToken, 'user_id': userId, 'email': email, 'name': name}),
    );
    if (res.statusCode != 200) throw Exception('apple auth failed');
    final j = jsonDecode(res.body);
    return AuthTokens(accessToken: j['access_token'] as String, refreshToken: j['refresh_token'] as String);
  }

  static Future<AuthTokens?> refreshTokens(String refreshToken) async {
    try {
      final res = await http.post(
        Uri.parse('$_base/v1/auth/refresh'),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'refresh_token': refreshToken}),
      );
      if (res.statusCode != 200) return null;
      final j = jsonDecode(res.body);
      return AuthTokens(accessToken: j['access_token'] as String, refreshToken: j['refresh_token'] as String);
    } catch (_) {
      return null;
    }
  }

  static Future<void> revokeToken(String token, String refreshToken) async {
    try {
      await http.post(
        Uri.parse('$_base/v1/auth/revoke'),
        headers: {'Authorization': 'Bearer $token', 'Content-Type': 'application/json'},
        body: jsonEncode({'refresh_token': refreshToken}),
      );
    } catch (_) {}
  }

  static Future<List<Challenge>> getChallenges() async {
    try {
      final res = await http.get(Uri.parse('$_base/v1/challenges')).timeout(const Duration(seconds: 5));
      if (res.statusCode == 200) {
        final List data = jsonDecode(res.body);
        return data.map((j) => Challenge.fromJson(j)).toList();
      }
    } catch (_) {}
    return _fallbackChallenges;
  }

  static Future<Map<String, dynamic>> uploadVideo({
    required String challengeId,
    required String videoPath,
    required String token,
    double? lat,
    double? lng,
    String? capturedAt,
    String robot = 'generic_bimanual',
    String consentVersion = '1.0',
  }) async {
    final req = http.MultipartRequest('POST', Uri.parse('$_base/v1/submit/$challengeId'))
      ..headers['Authorization'] = 'Bearer $token'
      ..fields['robot'] = robot
      ..fields['consent_version'] = consentVersion
      ..files.add(await http.MultipartFile.fromPath('video', videoPath));

    if (lat != null) req.fields['lat'] = lat.toString();
    if (lng != null) req.fields['lng'] = lng.toString();
    if (capturedAt != null) req.fields['captured_at'] = capturedAt;

    final streamed = await req.send();
    final res = await http.Response.fromStream(streamed);
    if (res.statusCode != 200) return {'error': res.body};
    return jsonDecode(res.body);
  }

  static Future<Map<String, dynamic>> getSubmission({required String token, required String id}) async {
    try {
      final res = await http.get(
        Uri.parse('$_base/v1/submissions/$id'),
        headers: {'Authorization': 'Bearer $token'},
      ).timeout(const Duration(seconds: 10));
      if (res.statusCode == 401) { onUnauthorized?.call(); return {}; }
      if (res.statusCode == 200) return jsonDecode(res.body);
    } catch (_) {}
    return {};
  }

  static Future<Map<String, dynamic>> getProfile({required String token}) async {
    try {
      final res = await http.get(
        Uri.parse('$_base/v1/profile'),
        headers: {'Authorization': 'Bearer $token'},
      ).timeout(const Duration(seconds: 10));
      if (res.statusCode == 401) { onUnauthorized?.call(); return {}; }
      if (res.statusCode != 200) return {};
      return jsonDecode(res.body);
    } catch (_) {
      return {};
    }
  }

  static Future<List<AppUser>> getLeaderboard() async {
    try {
      final res = await http.get(Uri.parse('$_base/v1/leaderboard')).timeout(const Duration(seconds: 10));
      if (res.statusCode != 200) return [];
      final List data = jsonDecode(res.body);
      return data.map((j) => AppUser.fromJson(j)).toList();
    } catch (_) {
      return [];
    }
  }

  static Future<List<Submission>> getUserSubmissions({required String token, int limit = 20, int offset = 0}) async {
    try {
      final res = await http.get(
        Uri.parse('$_base/v1/submissions?limit=$limit&offset=$offset'),
        headers: {'Authorization': 'Bearer $token'},
      ).timeout(const Duration(seconds: 10));
      if (res.statusCode == 401) { onUnauthorized?.call(); return []; }
      if (res.statusCode != 200) return [];
      final List data = jsonDecode(res.body);
      return data.map((j) => Submission.fromJson(j)).toList();
    } catch (_) {
      return [];
    }
  }

  static Future<List<Map<String, dynamic>>> getCreditHistory({required String token, int limit = 20, int offset = 0}) async {
    try {
      final res = await http.get(
        Uri.parse('$_base/v1/credits/history?limit=$limit&offset=$offset'),
        headers: {'Authorization': 'Bearer $token'},
      ).timeout(const Duration(seconds: 10));
      if (res.statusCode == 401) { onUnauthorized?.call(); return []; }
      if (res.statusCode != 200) return [];
      final List data = jsonDecode(res.body);
      return data.cast<Map<String, dynamic>>();
    } catch (_) {
      return [];
    }
  }

  static Future<Map<String, dynamic>> getUserStats({required String token}) async {
    try {
      final res = await http.get(
        Uri.parse('$_base/v1/profile/stats'),
        headers: {'Authorization': 'Bearer $token'},
      ).timeout(const Duration(seconds: 10));
      if (res.statusCode == 401) { onUnauthorized?.call(); return {}; }
      if (res.statusCode != 200) return {};
      return jsonDecode(res.body);
    } catch (_) {
      return {};
    }
  }

  static Stream<Map<String, dynamic>> streamSubmission({required String token, required String id}) async* {
    final client = http.Client();
    try {
      final req = http.Request('GET', Uri.parse('$_base/v1/submissions/$id/stream'));
      req.headers['Authorization'] = 'Bearer $token';
      req.headers['Accept'] = 'text/event-stream';
      final resp = await client.send(req);
      if (resp.statusCode != 200) return;

      final buffer = StringBuffer();
      await for (final chunk in resp.stream.transform(const Utf8Decoder())) {
        buffer.write(chunk);
        final text = buffer.toString();
        final lines = text.split('\n');
        buffer.clear();
        buffer.write(lines.last);

        for (final line in lines.sublist(0, lines.length - 1)) {
          if (line.startsWith('data: ')) {
            final data = line.substring(6).trim();
            if (data.isNotEmpty) {
              try {
                yield jsonDecode(data) as Map<String, dynamic>;
              } catch (_) {}
            }
          }
        }
      }
    } finally {
      client.close();
    }
  }

  static Future<Map<String, dynamic>> getReferral({required String token}) async {
    try {
      final res = await http.get(
        Uri.parse('$_base/v1/referral'),
        headers: {'Authorization': 'Bearer $token'},
      ).timeout(const Duration(seconds: 10));
      if (res.statusCode == 401) { onUnauthorized?.call(); return {}; }
      if (res.statusCode != 200) return {};
      return jsonDecode(res.body);
    } catch (_) {
      return {};
    }
  }

  static Future<bool> redeemReferral({required String token, required String code}) async {
    try {
      final res = await http.post(
        Uri.parse('$_base/v1/referral/redeem'),
        headers: {'Authorization': 'Bearer $token', 'Content-Type': 'application/json'},
        body: jsonEncode({'code': code}),
      ).timeout(const Duration(seconds: 10));
      if (res.statusCode == 401) { onUnauthorized?.call(); return false; }
      return res.statusCode == 200;
    } catch (_) {
      return false;
    }
  }
}
