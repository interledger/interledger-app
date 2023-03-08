import 'package:flutter/material.dart';
import 'package:watsonia/styles/colors.dart';

final theme = ThemeData(
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
);
