import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:watsonia/styles/colors.dart';

class Card2 extends StatelessWidget {
  const Card2({
    super.key,
    required this.children,
  });

  final List<Widget> children;

  @override
  Widget build(BuildContext context) {
    return Container(
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(20),
          color: TWColors.bgPage,
        ),
        width: double.infinity,
        margin: const EdgeInsets.only(bottom: 16),
        padding: const EdgeInsets.all(8),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: children,
        ));
  }
}

class CardContent2 extends StatelessWidget {
  const CardContent2({
    super.key,
    required this.children,
  });

  final List<Widget> children;

  @override
  Widget build(BuildContext context) {
    return Container(
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(20),
          color: TWColors.bgPage,
        ),
        width: double.infinity,
        margin: const EdgeInsets.only(bottom: 16),
        padding: const EdgeInsets.all(8),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: children,
        ));
  }
}

class CustomCard extends StatelessWidget {
  final List<Widget> children;
  final EdgeInsetsGeometry padding;
  final BorderRadiusGeometry borderRadius;
  final Color backgroundColor;

  const CustomCard({
    super.key,
    required this.children,
    this.padding = const EdgeInsets.all(8.0),
    this.borderRadius = const BorderRadius.all(Radius.circular(20.0)),
    this.backgroundColor = Colors.white,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(20),
          color: TWColors.bgPage,
        ),
        width: double.infinity,
        margin: const EdgeInsets.only(bottom: 16),
        padding: const EdgeInsets.all(8),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: children,
        ));
  }
}

class CardHeader extends StatelessWidget {
  final Widget child;
  final EdgeInsetsGeometry padding;

  const CardHeader({
    super.key,
    required this.child,
    this.padding = const EdgeInsets.fromLTRB(8.0, 8.0, 8.0, 0.0),
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: padding,
      child: child,
    );
  }
}

class CardTitle extends StatelessWidget {
  final String title;
  final TextStyle style;

  const CardTitle({
    super.key,
    required this.title,
    this.style = const TextStyle(
      color: TWColors.textStrong,
      fontWeight: FontWeight.w500,
      fontSize: 18,
    ),
  });

  @override
  Widget build(BuildContext context) {
    return Text(
      title,
      style: GoogleFonts.inter(
        textStyle: style,
      ),
    );
  }
}

class CardIcon extends StatelessWidget {
  final Widget icon;
  final EdgeInsetsGeometry padding;
  final Color backgroundColor;

  const CardIcon({
    super.key,
    required this.icon,
    this.padding = const EdgeInsets.all(20.0),
    this.backgroundColor = Colors.blue,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: padding,
      decoration: BoxDecoration(
        color: backgroundColor,
        shape: BoxShape.circle,
      ),
      child: icon,
    );
  }
}

class CardContent extends StatelessWidget {
  final Widget child;
  final EdgeInsetsGeometry padding;

  const CardContent({
    super.key,
    required this.child,
    this.padding = const EdgeInsets.all(8.0),
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: padding,
      child: child,
    );
  }
}

class CardLink extends StatelessWidget {
  final Widget child;
  final VoidCallback? onPressed;
  final Color backgroundColor;
  final Color hoverColor;
  final EdgeInsetsGeometry padding;

  const CardLink({
    super.key,
    required this.child,
    this.onPressed,
    this.backgroundColor = Colors.transparent,
    this.hoverColor = Colors.blue,
    this.padding = const EdgeInsets.all(12.0),
  });

  @override
  Widget build(BuildContext context) {
    return InkWell(
      splashColor: hoverColor,
      focusColor: TWColors.bgContainerHover,
      onTap: onPressed,
      child: Container(
        padding: padding,
        decoration: BoxDecoration(
          color: backgroundColor,
          borderRadius: BorderRadius.circular(12.0),
        ),
        child: child,
      ),
    );
  }
}

class CardButton extends StatelessWidget {
  final Widget child;
  final VoidCallback? onPressed;
  final Color backgroundColor;
  final Color hoverColor;
  final EdgeInsetsGeometry padding;

  const CardButton({
    super.key,
    required this.child,
    this.onPressed,
    this.backgroundColor = Colors.blue,
    this.hoverColor = Colors.blueAccent,
    this.padding = const EdgeInsets.all(12.0),
  });

  @override
  Widget build(BuildContext context) {
    return ElevatedButton(
      onPressed: onPressed,
      style: ElevatedButton.styleFrom(
        backgroundColor: backgroundColor,
        foregroundColor: hoverColor,
        padding: padding,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(16.0),
        ),
      ),
      child: child,
    );
  }
}
