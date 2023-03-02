import 'package:flutter/material.dart';
import 'package:watsonia/router.dart';
import 'package:watsonia/styles/colors.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  // This widget is the root of your application.
  @override
  Widget build(BuildContext context) {
    return MaterialApp.router(
      title: 'Fynbos',
      routerConfig: appRouter,
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        useMaterial3: true,
        colorScheme: const ColorScheme(
          background: TWColors.bgApp,
          brightness: Brightness.light,
          primary: TWColors.bgApp,
          onPrimary: TWColors.bgApp,
          secondary: TWColors.bgApp,
          onSecondary: TWColors.bgApp,
          error: TWColors.bgApp,
          onError: TWColors.bgApp,
          onBackground: TWColors.bgApp,
          surface: TWColors.bgApp,
          onSurface: TWColors.textStrong,
        ),
      ),
    );
  }
}
