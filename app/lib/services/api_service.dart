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

  static Future<bool> uploadVideo(String challengeId) async {
    final res = await http.post(
      Uri.parse('$_base/submit/$challengeId'),
      headers: {'Content-Type': 'application/json'},
    );
    return res.statusCode == 200;
  }

  static Future<Map<String, dynamic>> getProfile() async {
    final res = await http.get(Uri.parse('$_base/profile'));
    if (res.statusCode != 200) return {};
    return jsonDecode(res.body);
  }
}
