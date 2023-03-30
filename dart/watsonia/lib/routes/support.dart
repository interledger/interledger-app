import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:watsonia/components/floating_action_button.dart';
import 'package:watsonia/components/nav_drawer.dart';
import 'package:watsonia/styles/colors.dart';

class SupportRoute extends StatefulWidget {
  const SupportRoute({super.key});

  @override
  State<SupportRoute> createState() => _SupportRouteState();
}

class _SupportRouteState extends State<SupportRoute> {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      // We don't need an app bar on this page because the shell route scaffold will render it for us.
      appBar: AppBar(title: const Text("Support"),),
      body: ListView(
        padding: const EdgeInsets.fromLTRB(16, 0, 16, 0),
        children: <Widget>[
          Card(
            margin: const EdgeInsets.all(0),
            elevation: 0,
            color: TWColors.white,
            child: Padding(
              padding: const EdgeInsets.fromLTRB(16, 16, 16, 24),
              child: Text(
                'Transactions',
                style: GoogleFonts.poppins(
                  textStyle: const TextStyle(
                    fontWeight: FontWeight.w500,
                    fontSize: 24,
                  ),
                ),
              ),
            ),
          )
        ],
      ),
      drawerScrimColor: TWColors.bgScrim,
      drawer: const NavDrawer(),
      floatingActionButton: const PayFAB(),
    );
  }
}
