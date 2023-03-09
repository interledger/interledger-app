import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:watsonia/styles/colors.dart';

final theme = ThemeData(
  useMaterial3: true,
  appBarTheme: const AppBarTheme(
    backgroundColor: TWColors.bgApp,
    systemOverlayStyle: SystemUiOverlayStyle(
      statusBarColor: TWColors.bgApp,
      // For Android (dark icons)
      statusBarBrightness: Brightness.light,
      // For iOS (dark icons)
      statusBarIconBrightness: Brightness.dark,
      systemNavigationBarColor: TWColors.bgApp,
      systemNavigationBarDividerColor: TWColors.bgApp,
    ),
  ),
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
);
