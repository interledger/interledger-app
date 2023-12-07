import 'package:flutter/cupertino.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import 'package:watsonia/error.dart';
import 'package:watsonia/routes/home.dart';
import 'package:watsonia/routes/landing.dart';
import 'package:watsonia/routes/login/route.dart';
import 'package:watsonia/routes/pay.dart';
import 'package:watsonia/routes/settings.dart';
import 'package:watsonia/routes/settings.profile-public.dart';
import 'package:watsonia/routes/settings.profile-public.name/route.dart';
import 'package:watsonia/routes/signup.dart';
import 'package:watsonia/routes/support.dart';
import 'package:watsonia/routes/transactions.dart';
import 'package:watsonia/services/auth_service.dart';
import 'package:watsonia/stores/auth_store.dart';

final GlobalKey<NavigatorState> rootNavigatorKey =
    GlobalKey<NavigatorState>(debugLabel: 'root');

final appRouter = GoRouter(
  navigatorKey: rootNavigatorKey,
  // refreshListenable:,
  redirect: _authGuard,
  errorBuilder: (context, state) => const ErrorRoute(),
  routes: [
    // TODO Probably want to have one home page that does both like protea
    AppRoute('/', (_) => const HomeRoute(), true),
    AppRoute('/landing', (_) => const LandingRoute()),
    // TODO Then can put signup and login as children of home
    AppRoute('/signup', (_) => const SignupRoute()),
    AppRoute('/login', (_) => const LoginRoute()),
    AppRoute('/payments', (_) => const PaymentsRoute(), true),
    AppRoute('/settings', (_) => const SettingsRoute(), true, [
      AppRoute(
        'profile-public',
        (_) => const SettingsProfilePublicRoute(),
        false,
        [
          AppRoute('name', (_) => const SettingsProfilePublicNameRoute()),
        ],
      ),
    ]),
    AppRoute('/support', (_) => const SupportRoute(), true),
    // TODO Make pay redirect to Home and open the pay dialog?
    AppRoute('/pay', (_) => const PayRoute()),
  ],
);

String? _authGuard(BuildContext context, GoRouterState state) {
  // NOTE we use context.read here because we don't need to listen to the value
  final AuthStatus authStatus = context.read<AuthStore>().status;

  // Don't redirect if we haven't initialised auth yet
  // if (authStatus == AuthStatus.uninitialized) return null;

  final bool loggingIn = state.uri.toString() == '/login';
  final bool signingUp = state.uri.toString() == '/signup';
  final bool landing = state.uri.toString() == '/landing';

  // Go to / if the user is signed in and tries to go to /login.
  if (authStatus == AuthStatus.authenticated &&
      (loggingIn || signingUp || landing)) {
    return '/';
  }
  // Go to / if the user is signed in and tries to go to /signup.
  else if (authStatus == AuthStatus.authenticated && signingUp) {
    return '/';
  }
  // Go to /landing if the user is not signed in
  // TODO To remove the initial auth jank we need another state for when the user is not authenticated
  else if (authStatus == AuthStatus.uninitialized && !loggingIn && !signingUp) {
    return '/landing';
  }

  // else no redirect
  return null;
}

/// Syntactic sugar to make the router declaration easier to read.
/// Also, allows fade transition on home pages, and cupertino page transitions elsewhere.
class AppRoute extends GoRoute {
  AppRoute(
    String path,
    Widget Function(GoRouterState state) builder, [
    bool fade = false,
    List<GoRoute> routes = const [],
  ]) : super(
          parentNavigatorKey: rootNavigatorKey,
          path: path,
          routes: routes,
          pageBuilder: (context, state) {
            if (fade) {
              return CustomTransitionPage(
                key: state.pageKey,
                child: builder(state),
                transitionsBuilder:
                    (context, animation, secondaryAnimation, child) {
                  return FadeTransition(
                    opacity: CurveTween(curve: Curves.easeInOutCirc)
                        .animate(animation),
                    child: child,
                  );
                },
              );
            } else {
              return CupertinoPage(
                key: state.pageKey,
                child: builder(state),
              );
            }
          },
        );
}
