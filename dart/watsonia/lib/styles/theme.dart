import 'package:flutter/services.dart';
import 'package:watsonia/components.dart';
import 'package:watsonia/styles/colors.dart';

final theme = ThemeData(
    useMaterial3: true,
    appBarTheme: const AppBarTheme(
      backgroundColor: TWColors.transparent,
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
    textTheme: const TextTheme(),
    filledButtonTheme: FilledButtonThemeData(
      style: ButtonStyle(
        minimumSize: MaterialStateProperty.all(const Size(double.infinity, 48)),
        backgroundColor: MaterialStateProperty.resolveWith<Color?>(
          (Set<MaterialState> states) {
            if (states.contains(MaterialState.pressed)) {
              return TWColors.bgContainerPrimaryActive;
            } else if (states.contains(MaterialState.disabled)) {
              return TWColors.bgDisabled;
            }
            return TWColors.bgPrimary;
          },
        ),
      ),
    ),
    textButtonTheme: TextButtonThemeData(
      style: ButtonStyle(
        minimumSize: MaterialStateProperty.all(const Size(0, 0)),
        padding: MaterialStateProperty.all(const EdgeInsets.all(0)),
        foregroundColor: MaterialStateProperty.resolveWith<Color?>(
          (Set<MaterialState> states) {
            if (states.contains(MaterialState.pressed)) {
              return TWColors.bgContainerPrimaryActive;
            } else if (states.contains(MaterialState.disabled)) {
              return TWColors.bgDisabled;
            }
            return TWColors.textPrimary;
          },
        ),
      ),
    ),
    inputDecorationTheme: InputDecorationTheme(
      filled: false,
      errorStyle: const TextStyle(
        color: TWColors.textError,
        fontSize: 14,
        fontWeight: FontWeight.w400,
      ),
      contentPadding: const EdgeInsets.fromLTRB(16, 8, 16, 8),
      fillColor: TWColors.transparent,
      border: OutlineInputBorder(
        borderRadius: BorderRadius.circular(12),
        borderSide: const BorderSide(color: TWColors.borderBase, width: 2),
      ),
      enabledBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(12),
        borderSide: const BorderSide(color: TWColors.borderBase, width: 2),
      ),
      focusedBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(12),
        borderSide: const BorderSide(color: TWColors.borderFocus, width: 2),
      ),
      errorBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(12),
        borderSide: const BorderSide(color: TWColors.borderError, width: 2),
      ),
      focusedErrorBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(12),
        borderSide: const BorderSide(color: TWColors.borderError, width: 2),
      ),
    ));
