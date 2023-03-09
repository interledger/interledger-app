import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:watsonia/styles/colors.dart';

class IndexRoute extends StatefulWidget {
  const IndexRoute({super.key, required this.title});

  // This widget is the home page of your application. It is stateful, meaning
  // that it has a State object (defined below) that contains fields that affect
  // how it looks.

  // This class is the configuration for the state. It holds the values (in this
  // case the title) provided by the parent (in this case the App widget) and
  // used by the build method of the State. Fields in a Widget subclass are
  // always marked "final".

  final String title;

  @override
  State<IndexRoute> createState() => _IndexRouteState();
}

class _IndexRouteState extends State<IndexRoute> {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      // We don't need an app bar on this page because the ShellRoute Scaffold will render it for us.
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
                'Welcome Cairin',
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
    );
  }
}
