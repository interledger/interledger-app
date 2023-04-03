import type { InputHTMLAttributes } from 'react'
import { forwardRef } from 'react'
import { Router } from '~/components/Router'

interface TextFieldProps extends InputHTMLAttributes<HTMLInputElement> {
  // Override the `className` of the root `div` of the Input. Defaults to **min-w-full**.
  className?: string
  // The label value.
  label?: string
  // Appended text to the label (Footer notes)
  labelSuffix?: string
  // Text right aligned to the label.
  labelLink?: string
  // Where the labelLink should go.
  labelLinkTo?: string
  // The message from errors produced by form validation.
  errorMessage?: string
  // Any success messages produced by form validation.
  successMessage?: string

  // Prefix text to show in front of the input.
  prefix?: string
  prefixIcon?: JSX.Element
  appendIcon?: JSX.Element
}

export const TextField = forwardRef<any, TextFieldProps>(
  (
    {
      className,
      label,
      labelSuffix,
      labelLink,
      labelLinkTo,
      errorMessage,
      successMessage,
      prefix,
      prefixIcon,
      appendIcon,
      ...inputProps
    },
    ref
  ) => {
    return (
      <div className={className || 'min-w-full'}>
        <div className='flex justify-between'>
          <label
            htmlFor={inputProps.id}
            className='ml-2 block text-sm font-medium text-medium'
          >
            {label} {labelSuffix && <sup>{labelSuffix}</sup>}
          </label>
          {labelLink && labelLinkTo && (
            <Router
              to={labelLinkTo}
              className='mr-2 text-sm font-medium text-primary'
            >
              {labelLink}
            </Router>
          )}
        </div>
        <div className='mt-1 block h-12 w-full rounded-xl border-2 border-base focus-within:border-focus focus-within:ring-0'>
          <div className='flex h-full items-center justify-between overflow-hidden rounded-[10px]'>
            {prefixIcon && (
              <div className='-mr-4 flex h-full items-center px-3'>
                {prefixIcon}
              </div>
            )}
            {prefix && (
              <span className='z-10 -mr-3 ml-4 font-medium text-weak'>
                {prefix}
              </span>
            )}
            <input
              ref={ref}
              {...inputProps}
              className='z-0 h-full w-full overflow-hidden border-none bg-transparent px-4 focus:ring-0'
            />
            {appendIcon && (
              <div className='-ml-3 flex h-full items-center px-3'>
                {appendIcon}
              </div>
            )}
          </div>
        </div>
        <div className='h-7 pl-2 pt-2'>
          {errorMessage && <p className='text-sm text-error'>{errorMessage}</p>}
          {successMessage && !errorMessage && (
            <p className='text-sm text-success'>{successMessage}</p>
          )}
        </div>
      </div>
    )
  }
)

TextField.displayName = 'TextField'

// Convert this component to dart to be used in a flutter project.
// Path: dart/watsonia/lib/components/text_field.dart
// Compare this snippet from dart/watsonia/lib/components/card.dart:
// class Card extends StatelessWidget {
//   const Card({
//     super.key,
//     required this.children,
//   });
//
//   final List<Widget> children;
//
//   @override
//   Widget build(BuildContext context) {
//     return Container(
//         decoration: BoxDecoration(
//             borderRadius: BorderRadius.circular(20), color: Colors.white),
//         width: double.infinity,
//         margin: const EdgeInsets.fromLTRB(16, 0, 16, 16),
//         padding: const EdgeInsets.all(16),
//         child: Column(
//             crossAxisAlignment: CrossAxisAlignment.start, children: children));
//   }
// }
//
// class TextField extends StatelessWidget {
//   const TextField({
//     super.key,
//     required this.label,
//     required this.controller,
//     required this.hint,
//     this.obscureText = false,
//     this.autocorrect = false,
//     this.autofocus = false,
//     this.keyboardType,
//     this.validator,
//     this.onSaved,
//     this.onChanged,
//   });
//
//   final String label;
//   final TextEditingController controller;
//   final String hint;
//   final bool obscureText;
//   final bool autocorrect;
//   final bool autofocus;
//   final TextInputType? keyboardType;
//   final String? Function(String?)? validator;
//   final void Function(String?)? onSaved;
//   final void Function(String?)? onChanged;
//
//   @override
//   Widget build(BuildContext context) {
//     return Column(
//       crossAxisAlignment: CrossAxisAlignment.start,
//       children: [
//         Text(
//           label,
//           style: const TextStyle(
//             color: Colors.black,
//             fontSize: 16,
//             fontWeight: FontWeight.w600,
//           ),
//         ),
//         const SizedBox(height: 8),
//         TextFormField(
//           controller: controller,
//           obscureText: obscureText,
//           autocorrect: autocorrect,
//           autofocus: autofocus,
//           keyboardType: keyboardType,
//           validator: validator,
//           onSaved: onSaved,
//           onChanged: onChanged,
//           style: const TextStyle(
//             color: Colors.black,
//             fontSize: 16,
//           ),
//           decoration: InputDecoration(
//             hintText: hint,
//             hintStyle: const TextStyle(
//               color: Colors.grey,
//               fontSize: 16,
//             ),
//             contentPadding: const EdgeInsets.fromLTRB(20, 8, 20, 8),
//             enabledBorder: OutlineInputBorder(
//               borderSide: const BorderSide(color: Colors.grey, width: 1),
//               borderRadius: BorderRadius.circular(8),
//             ),
//             focusedBorder: OutlineInputBorder(
//               borderSide: const BorderSide(color: Colors.grey, width: 1),
//               borderRadius: BorderRadius.circular(8),
//             ),
//             errorBorder: OutlineInputBorder(
//               borderSide: const BorderSide(color: Colors.red, width: 1),
//               borderRadius: BorderRadius.circular(8),
//             ),
//             focusedErrorBorder: OutlineInputBorder(
//               borderSide: const BorderSide(color: Colors.red, width: 1),
//               borderRadius: BorderRadius.circular(8),
//             ),
//           ),
//         ),
//       ],
//     );
//   }
// }
//
