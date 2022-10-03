import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
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
      theme: ThemeData(
        // This is the theme of your application.
        //
        // Try running your application with "flutter run". You'll see the
        // application has a blue toolbar. Then, without quitting the app, try
        // changing the primarySwatch below to Colors.green and then invoke
        // "hot reload" (press "r" in the console where you ran "flutter run",
        // or simply save your changes to "hot reload" in a Flutter IDE).
        // Notice that the counter didn't reset back to zero; the application
        // is not restarted.
        primarySwatch: Colors.blue,
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
  int _counter = 0;

  void _incrementCounter() {
    setState(() {
      // This call to setState tells the Flutter framework that something has
      // changed in this State, which causes it to rerun the build method below
      // so that the display can reflect the updated values. If we changed
      // _counter without calling setState(), then the build method would not be
      // called again, and so nothing would appear to happen.
      _counter++;
    });
  }

  @override
  Widget build(BuildContext context) {
    // This method is rerun every time setState is called, for instance as done
    // by the _incrementCounter method above.
    //
    // The Flutter framework has been optimized to make rerunning build methods
    // fast, so that you can just rebuild anything that needs updating rather
    // than having to individually change instances of widgets.
    return Container(
        width: 100,
        color: const Color(0xffF8FAFC),
        child: SafeArea(
          child: Column(
            children: [
              Container(
                  height: 64,
                  width: double.infinity,
                  alignment: Alignment.centerLeft,
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.center,
                    children: [
                      const Padding(
                        padding: EdgeInsets.all(16.0),
                        child: Icon(Icons.menu),
                      ),
                      Text(
                        "fynbos",
                        style: GoogleFonts.poppins(
                          textStyle: const TextStyle(
                              color: Color(0xFF0F172A),
                              fontSize: 24,
                              fontWeight: FontWeight.w500),
                        ),
                      ),
                    ],
                  )),
              Container(
                decoration: BoxDecoration(
                    borderRadius: BorderRadius.circular(20),
                    color: Colors.white),
                width: double.infinity,
                margin: const EdgeInsets.fromLTRB(16, 0, 16, 16),
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      "Sign up",
                      style: GoogleFonts.poppins(
                        textStyle: const TextStyle(
                            color: Color(0xFF0F172A),
                            fontSize: 24,
                            fontWeight: FontWeight.w500),
                      ),
                    ),
                    Text(
                      "Here's what we will need to create your account.",
                      style: GoogleFonts.poppins(
                        textStyle: const TextStyle(
                            decoration: TextDecoration.none,
                            color: Color(0xFF0F172A),
                            fontSize: 16,
                            fontWeight: FontWeight.w400),
                      ),
                    ),
                    Row(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Container(
                          width: 32,
                          height: 32,
                          decoration: BoxDecoration(
                              color: TWColors.yellow[300],
                              borderRadius: const BorderRadius.only(
                                  topLeft: Radius.circular(32))),
                        ),
                        Container(
                          margin: const EdgeInsets.fromLTRB(0, 0, 16, 0),
                          width: 32,
                          height: 32,
                          decoration: BoxDecoration(
                            color: TWColors.slate[500],
                            borderRadius: const BorderRadius.only(
                              bottomLeft: Radius.circular(32),
                            ),
                          ),
                        ),
                        Flexible(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                "Personal details",
                                style: GoogleFonts.poppins(
                                  textStyle: const TextStyle(
                                      color: Color(0xFF0F172A),
                                      fontSize: 16,
                                      fontWeight: FontWeight.w500),
                                ),
                              ),
                              Text(
                                "Your legal name, email address and country of residence.",
                                style: GoogleFonts.poppins(
                                  textStyle: const TextStyle(
                                      color: Color(0xFF0F172A),
                                      fontSize: 12,
                                      fontWeight: FontWeight.w400),
                                ),
                              ),
                            ],
                          ),
                        ),
                      ],
                    ),
                    Row(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Container(
                          width: 32,
                          height: 32,
                          decoration: BoxDecoration(
                              color: TWColors.slate[300],
                              borderRadius: const BorderRadius.only(
                                  bottomLeft: Radius.circular(32))),
                        ),
                        Container(
                          margin: const EdgeInsets.fromLTRB(0, 0, 16, 0),
                          width: 32,
                          height: 32,
                          decoration: BoxDecoration(
                            color: TWColors.lime[500],
                            borderRadius: BorderRadius.circular(32),
                          ),
                        ),
                        Flexible(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                "Mobile phone number",
                                style: GoogleFonts.poppins(
                                  textStyle: const TextStyle(
                                      color: Color(0xFF0F172A),
                                      fontSize: 16,
                                      fontWeight: FontWeight.w500),
                                ),
                              ),
                              Text(
                                "A mobile phone number we can verify",
                                style: GoogleFonts.poppins(
                                  textStyle: const TextStyle(
                                      color: Color(0xFF0F172A),
                                      fontSize: 12,
                                      fontWeight: FontWeight.w400),
                                ),
                              ),
                            ],
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ],
          ),
        ));
  }
}
