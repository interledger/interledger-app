import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:watsonia/auth.dart';
import 'package:watsonia/router.dart';
import 'package:watsonia/styles/theme.dart';

import 'data/wallet.dart';

void main() {
  runApp(
    /// Providers are above [MyApp] instead of inside it, so that tests
    /// can use [MyApp] while mocking the providers
    MultiProvider(
      providers: [
        ChangeNotifierProvider(create: (_) => Auth()),
        ChangeNotifierProvider(create: (_) => Wallet()),
      ],
      child: const MyApp(),
    ),
  );
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp.router(
      title: 'Fynbos',
      routerConfig: appRouter,
      theme: theme,
      color: Colors.teal,
      debugShowCheckedModeBanner: false,
    );
  }
}
