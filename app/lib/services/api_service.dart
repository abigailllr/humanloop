import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/challenge.dart';
import '../models/submission.dart';
import '../models/user.dart';

class ApiService {
  static const _base = String.fromEnvironment('API_URL', defaultValue: 'http://localhost:8080');

  static Future<List<Challenge>> getChallenges() async {
    final res = await http.get(Uri.parse('$_base/v1/challenges'));
    if (res.statusCode != 200) return [];
    final List data = jsonDecode(res.body);
    return data.map((j) => Challenge.fromJson(j)).toList();
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

  static Future<Map<String, dynamic>> getProfile({required String token}) async {
    final res = await http.get(
      Uri.parse('$_base/v1/profile'),
      headers: {'Authorization': 'Bearer $token'},
    );
    if (res.statusCode != 200) return {};
    return jsonDecode(res.body);
  }

  static Future<List<AppUser>> getLeaderboard() async {
    final res = await http.get(Uri.parse('$_base/v1/leaderboard'));
    if (res.statusCode != 200) return [];
    final List data = jsonDecode(res.body);
    return data.map((j) => AppUser.fromJson(j)).toList();
  }

  static Future<List<Submission>> getUserSubmissions({required String token}) async {
    final res = await http.get(
      Uri.parse('$_base/v1/submissions'),
      headers: {'Authorization': 'Bearer $token'},
    );
    if (res.statusCode != 200) return [];
    final List data = jsonDecode(res.body);
    return data.map((j) => Submission.fromJson(j)).toList();
  }
}
