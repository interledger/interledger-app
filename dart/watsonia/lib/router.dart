import 'package:flutter/cupertino.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import 'package:watsonia/auth.dart';
import 'package:watsonia/routes/index.dart';
import 'package:watsonia/routes/pay.dart';
import 'package:watsonia/routes/settings.dart';
import 'package:watsonia/routes/signup.about.dart';
import 'package:watsonia/routes/signup.dart';
import 'package:watsonia/routes/support.dart';
import 'package:watsonia/routes/transactions.dart';

final GlobalKey<NavigatorState> _rootNavigatorKey =
    GlobalKey<NavigatorState>(debugLabel: 'root');

final appRouter = GoRouter(
  navigatorKey: _rootNavigatorKey,
  redirect: _authGuard,
  routes: [
    AppRoute('/', (_) => const IndexRoute(), true),
    AppRoute('/transactions', (_) => const TransactionsRoute(), true),
    AppRoute('/settings', (_) => const SettingsRoute(), true),
    AppRoute('/support', (_) => const SupportRoute(), true),
    AppRoute('/pay', (_) => const PayRoute()),
    AppRoute('/signup', (_) => const SignupRoute()),
    AppRoute('/signup/about', (_) => const SignupAboutRoute()),
  ],
);

String? _authGuard(BuildContext context, GoRouterState state) {
  final bool isUser = context.read<Auth>().isUser;
  final bool loggingIn = state.subloc == '/signup';

  // Go to /signin if the user is not signed in
  if (!isUser && !loggingIn) {
    // Probably need to figure out signup
    return '/signup/about';
  }
  // Go to / if the user is signed in and tries to go to /signin.
  // else if (isUser && loggingIn) {
  //   return '/';
  // }

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
          parentNavigatorKey: _rootNavigatorKey,
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

// TODO implement the following routes as AppRoutes
// "/connections"
// "/connections/:connectionId"
// "/connections/add-a-public-key"
// "/contacts"
// "/"
// "/linked-account/:type"
// "/linked-account/:type/almost-there"
// "/linked-account/bank/widget"
// "/linked-account/card/widget"
// "/login"
// "/login/challenge"
// "/logout"
// "/pay"
// "/pay/amount"
// "/pay/confirm"
// "/payment-pointer"
// "/personal-details"
// "/personal-details/about"
// "/personal-details/address"
// "/recovery"
// "/recovery/password"
// "/settings"
// "/settings/linked-accounts"
// "/settings/linked-accounts/:accountId"
// "/settings/password"
// "/settings/profile-contact"
// "/settings/profile-personal"
// "/settings/profile-public"
// "/settings/profile-public/name"
// "/signup"
// "/signup/about"
// "/signup/password"
// "/signup/phone"
// "/support"
// "/transaction/open_payments_incoming/:transactionId"
// "/transaction/open_payments_outgoing/:transactionId"
// "/transactions"
// "/verify"
// "/waitlist"
