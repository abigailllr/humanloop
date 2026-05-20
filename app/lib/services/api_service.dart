import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/challenge.dart';
import '../models/submission.dart';
import '../models/user.dart';

class ApiService {
  static const _base = String.fromEnvironment('API_URL', defaultValue: 'http://localhost:8080');
  static void Function()? onUnauthorized;

  static const _fallbackChallenges = [
    Challenge(id: 'c1', title: 'Pick & Place', description: 'Pick up any object from a table and place it into a box.', submissions: 0),
    Challenge(id: 'c2', title: 'Fold It', description: 'Fold a piece of cloth or paper in half.', submissions: 0),
    Challenge(id: 'c3', title: 'Sort & Stack', description: 'Sort 5 objects by size from smallest to largest.', submissions: 0, difficulty: 'Medium'),
  ];

  static Future<String> authGoogle(String idToken) async {
    final res = await http.post(
      Uri.parse('$_base/v1/auth/google'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({'id_token': idToken}),
    );
    if (res.statusCode != 200) throw Exception('google auth failed');
    return jsonDecode(res.body)['token'] as String;
  }

  static Future<String> authApple({
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
    return jsonDecode(res.body)['token'] as String;
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
  }) async {
    final req = http.MultipartRequest('POST', Uri.parse('$_base/v1/submit/$challengeId'))
      ..headers['Authorization'] = 'Bearer $token'
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

  static Future<List<Submission>> getUserSubmissions({required String token}) async {
    try {
      final res = await http.get(
        Uri.parse('$_base/v1/submissions'),
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
}
