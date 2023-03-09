import 'package:flutter/cupertino.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import 'package:watsonia/auth.dart';
import 'package:watsonia/components/floating_action_button.dart';
import 'package:watsonia/components/nav_drawer.dart';
import 'package:watsonia/routes/index.dart';
import 'package:watsonia/routes/pay.dart';
import 'package:watsonia/routes/signup.dart';
import 'package:watsonia/routes/transactions.dart';
import 'package:watsonia/styles/colors.dart';

// We need two Global keys to manage the state of the two navigators.
// One is used by the Shell route to enable us to have a global NavDrawer.
// The other is used for all other pages that don't have a drawer.
// This can be removed after https://github.com/flutter/flutter/issues/26954 is fixed.
final GlobalKey<NavigatorState> _rootNavigatorKey =
    GlobalKey<NavigatorState>(debugLabel: 'root');
final GlobalKey<NavigatorState> _shellNavigatorKey =
    GlobalKey<NavigatorState>(debugLabel: 'shell');

final appRouter = GoRouter(
  navigatorKey: _rootNavigatorKey,
  redirect: _authGuard,
  routes: _routes,
);

List<RouteBase> _routes = [
  AppRoute('/pay', (_) => const PayRoute(title: 'Flutter Demo Home Page')),
  AppRoute('/signup', (_) => const SignupRoute()),

  // Don't put routes in here unless you need the NavDrawer from that route.
  ShellRoute(
    navigatorKey: _shellNavigatorKey,
    routes: [
      GoRoute(
        path: '/',
        pageBuilder: (BuildContext context, GoRouterState state) {
          return const NoTransitionPage(
              child: IndexRoute(title: 'Flutter Demo Home Page'));
        },
      ),
      GoRoute(
        path: '/transactions',
        pageBuilder: (BuildContext context, GoRouterState state) {
          return const NoTransitionPage(
              child: TransactionsRoute(title: 'Flutter Demo Home Page'));
        },
      ),
    ],
    pageBuilder: (BuildContext context, GoRouterState state, Widget child) {
      return CupertinoPage<dynamic>(
        child: Scaffold(
          appBar: AppBar(
            title: Image.asset(
              'images/Logo.png',
              height: 32,
            ),
          ),
          body: child,
          drawerScrimColor: TWColors.bgScrim,
          drawer: const NavDrawer(),
          floatingActionButton: const PayFAB(),
        ),
      );
    },
  )
];

String? _authGuard(BuildContext context, GoRouterState state) {
  final bool isUser = context.read<Auth>().isUser;
  final bool loggingIn = state.subloc == '/signup';

  // Go to /signin if the user is not signed in
  if (!isUser && !loggingIn) {
    // Probably need to figure out signup
    return '/signup';
  }
  // Go to / if the user is signed in and tries to go to /signin.
  // else if (isUser && loggingIn) {
  //   return '/';
  // }

  // else no redirect
  return null;
}

/// Syntactic sugar to make the router declaration easier to read.
class AppRoute extends GoRoute {
  AppRoute(
    String path,
    Widget Function(GoRouterState state) builder, {
    List<GoRoute> routes = const [],
  }) : super(
          parentNavigatorKey: _rootNavigatorKey,
          path: path,
          routes: routes,
          pageBuilder: (context, state) {
            return CupertinoPage(
              child: builder(state),
            );
          },
        );
}
