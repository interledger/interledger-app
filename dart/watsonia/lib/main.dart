import 'package:flutter/material.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:provider/provider.dart';
import 'package:watsonia/router.dart';
import 'package:watsonia/services/auth_service.dart';
import 'package:watsonia/storage.dart';
import 'package:watsonia/stores/auth_store.dart';
import 'package:watsonia/stores/kyc_store.dart';
import 'package:watsonia/stores/wallet_store.dart';
import 'package:watsonia/styles/colors.dart';
import 'package:watsonia/styles/theme.dart';

Future<void> main() async {
  await dotenv.load(fileName: '.env');
  // These singletons don't need to be passed as providers because they're not needed in the widget tree.
  final secureStorage = SecureStorage();
  final authService = AuthService();

  runApp(
    /// Providers are above [MyApp] instead of inside it, so that tests
    /// can use [MyApp] while mocking the providers
    MultiProvider(
      providers: [
        Provider<AuthStore>(
            create: (_) => AuthStore(authService, secureStorage)..init()),
        Provider<KYCStore>(create: (_) => KYCStore()),
        Provider<WalletStore>(create: (_) => WalletStore()..init()),
      ],
      child: const MyApp(), // LfSGlVD@nnKdPjh.info
    ),
  );
}

final GlobalKey<ScaffoldMessengerState> snackbarKey =
    GlobalKey<ScaffoldMessengerState>();

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp.router(
      title: 'Fynbos',
      routerConfig: appRouter,
      scaffoldMessengerKey: snackbarKey,
      theme: theme,
      color: TWColors.bgPrimary,
      debugShowCheckedModeBanner: false,
    );
  }
}
