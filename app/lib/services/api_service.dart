import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/challenge.dart';

class ApiService {
  static const _base = String.fromEnvironment('API_URL', defaultValue: 'http://localhost:8080');

  static Future<List<Challenge>> getChallenges() async {
    final res = await http.get(Uri.parse('$_base/challenges'));
    if (res.statusCode != 200) return [];
    final List data = jsonDecode(res.body);
    return data.map((j) => Challenge.fromJson(j)).toList();
  }

  static Future<Map<String, dynamic>> uploadVideo({
    required String challengeId,
    required String videoPath,
    required String token,
  }) async {
    final req = http.MultipartRequest('POST', Uri.parse('$_base/submit/$challengeId'))
      ..headers['Authorization'] = 'Bearer $token'
      ..files.add(await http.MultipartFile.fromPath('video', videoPath));

    final streamed = await req.send();
    final res = await http.Response.fromStream(streamed);
    if (res.statusCode != 200) return {'error': res.body};
    return jsonDecode(res.body);
  }

  static Future<Map<String, dynamic>> getProfile({required String token}) async {
    final res = await http.get(
      Uri.parse('$_base/profile'),
      headers: {'Authorization': 'Bearer $token'},
    );
    if (res.statusCode != 200) return {};
    return jsonDecode(res.body);
  }
}
