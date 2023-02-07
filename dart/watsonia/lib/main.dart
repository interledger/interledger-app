import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:watsonia/styles/colors.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  // This widget is the root of your application.
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Flutter Demo',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        useMaterial3: true,
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
      ),
      home: const MyHomePage(title: 'Flutter Demo Home Page'),
    );
  }
}

class MyHomePage extends StatefulWidget {
  const MyHomePage({super.key, required this.title});

  // This widget is the home page of your application. It is stateful, meaning
  // that it has a State object (defined below) that contains fields that affect
  // how it looks.

  // This class is the configuration for the state. It holds the values (in this
  // case the title) provided by the parent (in this case the App widget) and
  // used by the build method of the State. Fields in a Widget subclass are
  // always marked "final".

  final String title;

  @override
  State<MyHomePage> createState() => _MyHomePageState();
}

class _MyHomePageState extends State<MyHomePage> {
  int _count = 0;

  @override
  Widget build(BuildContext context) {
    // This method is rerun every time setState is called, for instance as done
    // by the _incrementCounter method above.
    //
    // The Flutter framework has been optimized to make rerunning build methods
    // fast, so that you can just rebuild anything that needs updating rather
    // than having to individually change instances of widgets.
    // return const SignupPage();
    return Scaffold(
      appBar: AppBar(
        title: Image.asset(
          'images/Logo.png',
          height: 32,
        ),
      ),
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
      floatingActionButton: FloatingActionButton.large(
        backgroundColor: TWColors.blue[500],
        foregroundColor: TWColors.white,
        onPressed: () => setState(() => _count++),
        tooltip: 'Increment Counter',
        child: const Icon(Icons.attach_money_outlined),
      ),
    );
  }
}
