import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:watsonia/styles/colors.dart';

class NavDrawer extends StatelessWidget {
  const NavDrawer({
    super.key,
  });

  @override
  Widget build(BuildContext context) {
    return Drawer(
      backgroundColor: TWColors.bgApp,
      width: 250,
      shape: const Border(),
      child: SafeArea(
        child: ListView(
          children: <Widget>[
            AppBar(
              leading: IconButton(
                icon: const Icon(Icons.menu_open_outlined),
                onPressed: () {
                  Navigator.pop(context);
                },
              ),
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
                selected: GoRouterState.of(context).location == '/',
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
                selected: GoRouterState.of(context).location == '/transactions',
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
                selected: GoRouterState.of(context).location == '/contact',
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
                selected: GoRouterState.of(context).location == '/settings',
                shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(20)),
                selectedTileColor: TWColors.bgContainerHover,
                selectedColor: TWColors.textStrong,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
