// TODO Probably want to abstract the material package so we don't depend on it too much?
import 'package:flutter/material.dart' as material;
import 'package:flutter/widgets.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:watsonia/styles/colors.dart';
import 'package:http/http.dart' as http;

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  // This widget is the root of your application.
  @override
  Widget build(BuildContext context) {
    return material.MaterialApp(
      title: 'Fynbos',
      debugShowCheckedModeBanner: false,
      theme: material.ThemeData(
        useMaterial3: true,
        colorScheme: const material.ColorScheme(
          background: TWColors.bgApp,
          brightness: material.Brightness.light,
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

Future<void> getSession() async {
  // const String basePath = r'http://cove-athletic-reed-scanning.trycloudflare.com';

  var url = Uri.http('cove-athletic-reed-scanning.trycloudflare.com', 'sessions/whoami');
  var response = await http.get(url);
  print('Response status: ${response.statusCode}');
  print('Response body: ${response.body}');
}

class _MyHomePageState extends State<MyHomePage> {
  int _count = 0;

  @override
  Widget build(BuildContext context) {
    getSession();
    // This method is rerun every time setState is called, for instance as done
    // by the _incrementCounter method above.
    //
    // The Flutter framework has been optimized to make rerunning build methods
    // fast, so that you can just rebuild anything that needs updating rather
    // than having to individually change instances of widgets.
    // return const SignupPage();
    return material.Scaffold(
      appBar: material.AppBar(
        title: Image.asset(
          'images/Logo.png',
          height: 32,
        ),
      ),
      body: ListView(
        padding: const EdgeInsets.fromLTRB(16, 0, 16, 0),
        children: <Widget>[
          material.Card(
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
      drawer: material.Drawer(
        width: 250,
        shape: const Border(),
        child: SafeArea(
          child: ListView(
            // padding: const EdgeInsets.fromLTRB(12, 16, 12, 16),
            children: <Widget>[
              material.AppBar(
                leading: const Icon(material.Icons.menu_open_outlined),
                title: Image.asset(
                  'images/Logo.png',
                  height: 32,
                ),
              ),
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 32, 16, 0),
                child: material.ListTile(
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
                child: material.ListTile(
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
                child: material.ListTile(
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
                child: material.ListTile(
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
      floatingActionButton: material.FloatingActionButton.large(
        backgroundColor: TWColors.blue[500],
        foregroundColor: TWColors.white,
        onPressed: () => setState(() => _count++),
        tooltip: 'Increment Counter',
        child: const Icon(material.Icons.attach_money_outlined),
      ),
    );
  }
}
