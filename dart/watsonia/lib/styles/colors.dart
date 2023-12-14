import 'package:flutter/painting.dart';

class TWColors {
  // This class is not meant to be instantiated or extended; this constructor
  // prevents instantiation and extension.
  TWColors._();

  // Color tokens

  // TODO: Should probably just replace these with values directly in the theme
  static const Color textStrong = slate900;
  static const Color textMedium = slate600;
  static const Color textWeak = slate500;
  static const Color textDisabled = slate400;
  static const Color textPrimary = blue500;
  static const Color textError = red500;
  static const Color textSuccess = green500;
  static const Color textOnColor = white;
  static const Color textInverted = white;

// TODO Update these tokens to match the designs
  static const Color bgApp = Color(0xfff8fafc);
  static const Color bgPage = white;
  static const Color bgContainer = Color(0xfff1f5f9);
  static const Color bgContainerHover = Color(0xffe2e8f0);
  static const Color bgStrong = Color(0xff94a3b8);
  static const Color bgDisabled = Color(0xffe2e8f0);
  static const Color bgPrimary = Color(0xff3b82f6);
  static const Color bgContainerPrimary = Color(0x00000000);
  static const Color bgContainerPrimaryHover = Color(0x00000000);
  static const Color bgContainerPrimaryActive = Color(0xff2563EB);
  static const Color bgScrim = Color(0xCC475569);
  static const Color bgSnackbar = Color(0x00000000);

  static const Color borderBase = Color(0xffCBD5E1);
  static const Color borderFocus = Color(0xff2563EB);
  static const Color borderHover = Color(0xff2563EB);
  static const Color borderActive = Color(0xff2563EB);
  static const Color borderError = Color(0xffb91c1c);

  // Tailwind colour styles

  static const Color transparent = Color(0x00000000);
  static const Color black = Color(0xFF000000);
  static const Color white = Color(0xFFFFFFFF);

  static const Color slate50 = Color(0xfff8fafc);
  static const Color slate100 = Color(0xfff1f5f9);
  static const Color slate200 = Color(0xffe2e8f0);
  static const Color slate300 = Color(0xffcbd5e1);
  static const Color slate400 = Color(0xff94a3b8);
  static const Color slate500 = Color(0xff64748b);
  static const Color slate600 = Color(0xff475569);
  static const Color slate700 = Color(0xff334155);
  static const Color slate800 = Color(0xff1e293b);
  static const Color slate900 = Color(0xff0f172a);
  static const Color slate950 = Color(0xff020617);

  static const Color red50 = Color(0xfffef2f2);
  static const Color red100 = Color(0xfffee2e2);
  static const Color red200 = Color(0xfffecaca);
  static const Color red300 = Color(0xfffca5a5);
  static const Color red400 = Color(0xfff87171);
  static const Color red500 = Color(0xffef4444);
  static const Color red600 = Color(0xffdc2626);
  static const Color red700 = Color(0xffb91c1c);
  static const Color red800 = Color(0xff991b1b);
  static const Color red900 = Color(0xff7f1d1d);
  static const Color red950 = Color(0xff450A0A);

  static const Color orange50 = Color(0xfffff7ed);
  static const Color orange100 = Color(0xffffedd5);
  static const Color orange200 = Color(0xfffed7aa);
  static const Color orange300 = Color(0xfffdba74);
  static const Color orange400 = Color(0xfffb923c);
  static const Color orange500 = Color(0xfff97316);
  static const Color orange600 = Color(0xffea580c);
  static const Color orange700 = Color(0xffc2410c);
  static const Color orange800 = Color(0xff9a3412);
  static const Color orange900 = Color(0xff7c2d12);
  static const Color orange950 = Color(0xff431407);

  static const Color yellow50 = Color(0xfffefce8);
  static const Color yellow100 = Color(0xfffef9c3);
  static const Color yellow200 = Color(0xfffef08a);
  static const Color yellow300 = Color(0xfffde047);
  static const Color yellow400 = Color(0xfffacc15);
  static const Color yellow500 = Color(0xffeab308);
  static const Color yellow600 = Color(0xffca8a04);
  static const Color yellow700 = Color(0xffa16207);
  static const Color yellow800 = Color(0xff854d0e);
  static const Color yellow900 = Color(0xff713f12);
  static const Color yellow950 = Color(0xff422006);

  static const Color lime50 = Color(0xfff7fee7);
  static const Color lime100 = Color(0xffecfccb);
  static const Color lime200 = Color(0xffd9f99d);
  static const Color lime300 = Color(0xffbef264);
  static const Color lime400 = Color(0xffa3e635);
  static const Color lime500 = Color(0xff84cc16);
  static const Color lime600 = Color(0xff65a30d);
  static const Color lime700 = Color(0xff4d7c0f);
  static const Color lime800 = Color(0xff3f6212);
  static const Color lime900 = Color(0xff365314);
  static const Color lime950 = Color(0xff1A2E05);

  static const Color green50 = Color(0xfff0fdf4);
  static const Color green100 = Color(0xffdcfce7);
  static const Color green200 = Color(0xffbbf7d0);
  static const Color green300 = Color(0xff86efac);
  static const Color green400 = Color(0xff4ade80);
  static const Color green500 = Color(0xff22c55e);
  static const Color green600 = Color(0xff16a34a);
  static const Color green700 = Color(0xff15803d);
  static const Color green800 = Color(0xff166534);
  static const Color green900 = Color(0xff14532d);
  static const Color green950 = Color(0xff052E16);

  static const Color sky50 = Color(0xfff0f9ff);
  static const Color sky100 = Color(0xffe0f2fe);
  static const Color sky200 = Color(0xffbae6fd);
  static const Color sky300 = Color(0xff7dd3fc);
  static const Color sky400 = Color(0xff38bdf8);
  static const Color sky500 = Color(0xff0ea5e9);
  static const Color sky600 = Color(0xff0284c7);
  static const Color sky700 = Color(0xff0369a1);
  static const Color sky800 = Color(0xff075985);
  static const Color sky900 = Color(0xff0c4a6e);
  static const Color sky950 = Color(0xff082F49);

  static const Color blue50 = Color(0xffeff6ff);
  static const Color blue100 = Color(0xffdbeafe);
  static const Color blue200 = Color(0xffbfdbfe);
  static const Color blue300 = Color(0xff93c5fd);
  static const Color blue400 = Color(0xff60a5fa);
  static const Color blue500 = Color(0xff3b82f6);
  static const Color blue600 = Color(0xff2563eb);
  static const Color blue700 = Color(0xff1d4ed8);
  static const Color blue800 = Color(0xff1e40af);
  static const Color blue900 = Color(0xff1e3a8a);
  static const Color blue950 = Color(0xff172554);

  static const Color indigo50 = Color(0xffeef2ff);
  static const Color indigo100 = Color(0xffe0e7ff);
  static const Color indigo200 = Color(0xffc7d2fe);
  static const Color indigo300 = Color(0xffa5b4fc);
  static const Color indigo400 = Color(0xff818cf8);
  static const Color indigo500 = Color(0xff6366f1);
  static const Color indigo600 = Color(0xff4f46e5);
  static const Color indigo700 = Color(0xff4338ca);
  static const Color indigo800 = Color(0xff3730a3);
  static const Color indigo900 = Color(0xff312e81);
  static const Color indigo950 = Color(0xff1E1B4B);

  static const Color purple50 = Color(0xfffaf5ff);
  static const Color purple100 = Color(0xfff3e8ff);
  static const Color purple200 = Color(0xffe9d5ff);
  static const Color purple300 = Color(0xffd8b4fe);
  static const Color purple400 = Color(0xffc084fc);
  static const Color purple500 = Color(0xffa855f7);
  static const Color purple600 = Color(0xff9333ea);
  static const Color purple700 = Color(0xff7e22ce);
  static const Color purple800 = Color(0xff6b21a8);
  static const Color purple900 = Color(0xff581c87);
  static const Color purple950 = Color(0xff3B0764);

  static const Color rose50 = Color(0xfffff1f2);
  static const Color rose100 = Color(0xffffe4e6);
  static const Color rose200 = Color(0xfffecdd3);
  static const Color rose300 = Color(0xfffda4af);
  static const Color rose400 = Color(0xfffb7185);
  static const Color rose500 = Color(0xfff43f5e);
  static const Color rose600 = Color(0xffe11d48);
  static const Color rose700 = Color(0xffbe123c);
  static const Color rose800 = Color(0xff9f1239);
  static const Color rose900 = Color(0xff881337);
  static const Color rose950 = Color(0xff4C0519);
}
