import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:provider/provider.dart';
import 'package:watsonia/auth.dart';
import 'package:watsonia/routes/index.dart';
import 'package:watsonia/routes/pay.dart';
import 'package:watsonia/routes/signup.dart';
import 'package:watsonia/routes/transactions.dart';
import 'package:watsonia/styles/colors.dart';

// TODO Look into whether page transitions can be made more normative
// https://github.com/flutter/packages/blob/main/packages/go_router/example/lib/transition_animations.dart

final appRouter = GoRouter(
  redirect: _handleRedirect,
  routes: _shellRoutes,
  observers: [HeroController()],
);

List<RouteBase> _routes = [
  GoRoute(
    path: '/',
    pageBuilder: (BuildContext context, GoRouterState state) {
      return const MaterialPage(
          child: MyHomePage(title: 'Flutter Demo Home Page'));
    },
  ),
  GoRoute(
    path: '/transactions',
    pageBuilder: (BuildContext context, GoRouterState state) {
      return const MaterialPage(
          child: TransactionsRoute(title: 'Flutter Demo Home Page'));
    },
  ),
  GoRoute(
    path: '/signup',
    pageBuilder: (BuildContext context, GoRouterState state) {
      return const MaterialPage(
          child: SignupPage());
    },
  ),
  GoRoute(
    path: '/pay',
    pageBuilder: (BuildContext context, GoRouterState state) {
      // print(state.fullpath);
      return const MaterialPage(
          child: PayPage(title: 'Flutter Demo Home Page'));
    },
  ),
];

List<RouteBase> _shellRoutes = [
  ShellRoute(
      routes: _routes,
      builder: (BuildContext context, GoRouterState state, Widget child) {
        print(GoRouterState.of(context).location);
        return Scaffold(
          appBar: AppBar(
            title: Image.asset(
              'images/Logo.png',
              height: 32,
            ),
          ),
          body: child,
          drawerScrimColor: TWColors.bgScrim,
          drawer: Drawer(
            width: 250,
            shape: const Border(),
            child: SafeArea(
              child: ListView(
                // padding: const EdgeInsets.fromLTRB(12, 16, 12, 16),
                children: <Widget>[
                  AppBar(
                    leading: const Icon(Icons.menu_open_outlined),
                    title: Image.asset(
                      'images/Logo.png',
                      height: 32,
                    ),
                  ),
                  Padding(
                    padding: const EdgeInsets.fromLTRB(16, 32, 16, 0),
                    child: ListTile(
                      onTap: () {
                        context.go('/');
                        Navigator.of(context).pop();
                      },
                      title: Text(
                        'Home',
                        style: GoogleFonts.poppins(
                          textStyle: const TextStyle(
                            fontSize: 16,
                          ),
                        ),
                      ),
                      selected: true,
                      shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(20)),
                      selectedTileColor: TWColors.bgContainerHover,
                      selectedColor: TWColors.textStrong,
                    ),
                  ),
                  Padding(
                    padding: const EdgeInsets.fromLTRB(16, 16, 16, 0),
                    child: ListTile(
                      onTap: () {
                        context.go('/transactions');
                        Navigator.of(context).pop();
                      },
                      title: Text(
                        'Transactions',
                        style: GoogleFonts.poppins(
                          textStyle: const TextStyle(
                            fontSize: 16,
                          ),
                        ),
                      ),
                      shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(20)),
                      selectedTileColor: TWColors.bgContainerHover,
                      selectedColor: TWColors.textStrong,
                    ),
                  ),
                  Padding(
                    padding: const EdgeInsets.fromLTRB(16, 16, 16, 0),
                    child: ListTile(
                      title: Text(
                        'Settings',
                        style: GoogleFonts.poppins(
                          textStyle: const TextStyle(
                            fontSize: 16,
                          ),
                        ),
                      ),
                      shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(20)),
                      selectedTileColor: TWColors.bgContainerHover,
                      selectedColor: TWColors.textStrong,
                    ),
                  ),
                  Padding(
                    padding: const EdgeInsets.fromLTRB(16, 16, 16, 0),
                    child: ListTile(
                      title: Text(
                        'Contact',
                        style: GoogleFonts.poppins(
                          textStyle: const TextStyle(
                            fontSize: 16,
                          ),
                        ),
                      ),
                      shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(20)),
                      selectedTileColor: TWColors.bgContainerHover,
                      selectedColor: TWColors.textStrong,
                    ),
                  ),
                ],
              ),
            ),
          ),
          floatingActionButton: state.location == '/' || state.location == '/pay'
              ? FloatingActionButton.large(
                  backgroundColor: TWColors.blue[500],
                  foregroundColor: TWColors.white,
                  // Push goes nested, go is for replacing root pages
                  onPressed: () => context.push('/pay'),
                  // onPressed: () => getSession(),
                  tooltip: 'Pay',
                  child: const Icon(Icons.attach_money_outlined),
                )
              : null,
        );
      })
];

String? _handleRedirect(BuildContext context, GoRouterState state) {
  final bool isUser = context.read<Auth>().isUser;
  final bool loggingIn = state.subloc == '/signup';

  // Go to /signin if the user is not signed in
  if (!isUser && !loggingIn) { // Probably need to figure out signup
    return '/signup';
  }
  // Go to / if the user is signed in and tries to go to /signin.
  else if (isUser && loggingIn) {
    return '/';
  }

  // else no redirect
  return null;
}

bool showFAB(GoRouterState state) {
  return state.location == '/' || state.location == '/transactions';
}
//
// final GoRouter _router = GoRouter(
//   navigatorKey: _rootNavigatorKey,
//   initialLocation: '/a',
//   routes: <RouteBase>[
//     /// Application shell
//     ShellRoute(
//       navigatorKey: _shellNavigatorKey,
//       builder: (BuildContext context, GoRouterState state, Widget child) {
//         return ScaffoldWithNavBar(child: child);
//       },
//       routes: <RouteBase>[
//         /// The first screen to display in the bottom navigation bar.
//         GoRoute(
//           path: '/a',
//           builder: (BuildContext context, GoRouterState state) {
//             return const ScreenA();
//           },
//           routes: <RouteBase>[
//             // The details screen to display stacked on the inner Navigator.
//             // This will cover screen A but not the application shell.
//             GoRoute(
//               path: 'details',
//               builder: (BuildContext context, GoRouterState state) {
//                 return const DetailsScreen(label: 'A');
//               },
//             ),
//           ],
//         ),
//
//         /// Displayed when the second item in the the bottom navigation bar is
//         /// selected.
//         GoRoute(
//           path: '/b',
//           builder: (BuildContext context, GoRouterState state) {
//             return const ScreenB();
//           },
//           routes: <RouteBase>[
//             /// Same as "/a/details", but displayed on the root Navigator by
//             /// specifying [parentNavigatorKey]. This will cover both screen B
//             /// and the application shell.
//             GoRoute(
//               path: 'details',
//               parentNavigatorKey: _rootNavigatorKey,
//               builder: (BuildContext context, GoRouterState state) {
//                 return const DetailsScreen(label: 'B');
//               },
//             ),
//           ],
//         ),
//
//         /// The third screen to display in the bottom navigation bar.
//         GoRoute(
//           path: '/c',
//           builder: (BuildContext context, GoRouterState state) {
//             return const ScreenC();
//           },
//           routes: <RouteBase>[
//             // The details screen to display stacked on the inner Navigator.
//             // This will cover screen A but not the application shell.
//             GoRoute(
//               path: 'details',
//               builder: (BuildContext context, GoRouterState state) {
//                 return const DetailsScreen(label: 'C');
//               },
//             ),
//           ],
//         ),
//       ],
//     ),
//   ],
// );
