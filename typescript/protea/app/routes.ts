import {
  UNSAFE_flatRoutes as flatRoutes,
  UNSAFE_getRouteConfigAppDirectory as getAppDirectory,
  UNSAFE_routeManifestToRouteConfig as routeManifestToRouteConfig
} from '@remix-run/dev'

export default routeManifestToRouteConfig(
  flatRoutes(getAppDirectory(), ['**/*.stories.tsx', '**/*.test.{ts,tsx}'])
)
