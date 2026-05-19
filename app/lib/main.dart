import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'theme/colors.dart';
import 'screens/onboarding_screen.dart';
import 'screens/login_screen.dart';
import 'screens/feed_screen.dart';
import 'screens/camera_screen.dart';
import 'screens/leaderboard_screen.dart';
import 'screens/history_screen.dart';
import 'screens/profile_screen.dart';

late SharedPreferences prefs;

final prefsProvider = Provider<SharedPreferences>((_) => prefs);

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  SystemChrome.setSystemUIOverlayStyle(const SystemUiOverlayStyle(
    statusBarColor: Colors.transparent,
    statusBarIconBrightness: Brightness.dark,
  ));
  prefs = await SharedPreferences.getInstance();
  runApp(const ProviderScope(child: HumanLoopApp()));
}

class HumanLoopApp extends StatelessWidget {
  const HumanLoopApp({super.key});

  @override
  Widget build(BuildContext context) {
    final hasOnboarded = prefs.getBool('onboarded') ?? false;
    final hasToken = prefs.getString('auth_token') != null;

    Widget home;
    if (hasToken) {
      home = const RootNav();
    } else if (hasOnboarded) {
      home = const LoginScreen();
    } else {
      home = const OnboardingScreen();
    }

    return MaterialApp(
      title: 'HumanLoop',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        useMaterial3: true,
        fontFamily: 'SF Pro Display',
        colorScheme: const ColorScheme.light(
          primary: AppColors.primary,
          surface: AppColors.surfaceGray,
          onSurface: AppColors.textPrimary,
        ),
        scaffoldBackgroundColor: AppColors.background,
      ),
      home: home,
    );
  }
}

class RootNav extends StatefulWidget {
  const RootNav({super.key});

  @override
  State<RootNav> createState() => _RootNavState();
}

class _RootNavState extends State<RootNav> {
  int _index = 0;

  static const _screens = [
    FeedScreen(),
    CameraScreen(),
    LeaderboardScreen(),
    HistoryScreen(),
    ProfileScreen(),
  ];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: _screens[_index],
      bottomNavigationBar: Container(
        decoration: const BoxDecoration(
          border: Border(top: BorderSide(color: AppColors.border, width: 1)),
        ),
        child: NavigationBar(
          backgroundColor: Colors.white,
          surfaceTintColor: Colors.transparent,
          indicatorColor: AppColors.primaryIndicator,
          selectedIndex: _index,
          onDestinationSelected: (i) => setState(() => _index = i),
          destinations: const [
            NavigationDestination(
              icon: Icon(Icons.bolt_outlined),
              selectedIcon: Icon(Icons.bolt, color: AppColors.primary),
              label: 'Challenges',
            ),
            NavigationDestination(
              icon: Icon(Icons.videocam_outlined),
              selectedIcon: Icon(Icons.videocam, color: AppColors.primary),
              label: 'Film',
            ),
            NavigationDestination(
              icon: Icon(Icons.leaderboard_outlined),
              selectedIcon: Icon(Icons.leaderboard, color: AppColors.primary),
              label: 'Leaders',
            ),
            NavigationDestination(
              icon: Icon(Icons.history_outlined),
              selectedIcon: Icon(Icons.history, color: AppColors.primary),
              label: 'History',
            ),
            NavigationDestination(
              icon: Icon(Icons.person_outline),
              selectedIcon: Icon(Icons.person, color: AppColors.primary),
              label: 'Profile',
            ),
          ],
        ),
      ),
    );
  }
}
