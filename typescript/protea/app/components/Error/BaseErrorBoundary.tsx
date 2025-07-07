import React from "react";
import { isRouteErrorResponse, useRouteError, type ErrorResponse } from "@remix-run/react";

export type BaseErrorBoundaryProps = {
  routeErrorConfig?: {
    fn?: (error: ErrorResponse) => void;
    render: (error: ErrorResponse) => React.ReactElement;
  }
  unexpectedErrorConfig?: {
    fn?: (error: Error) => void;
    render: (error: Error) => React.ReactElement;
  }
}

const defaultErrorBoundaryProps: Required<BaseErrorBoundaryProps> = {
  routeErrorConfig: {
    fn: (error: ErrorResponse) => {
      console.error("Caught route error:", error);
    },
    render: (error: ErrorResponse) => (
    <div style={{ color: "red", padding: "2rem", textAlign: "center" }}>
      <h1>An error occurred</h1>
      <p>
        {error.status} {error.statusText}
      </p>
      </div>
    )
  },
  unexpectedErrorConfig: {
    fn: (error: Error) => {
      console.error("Uncaught error:", error);
    },
    render: (error: Error) => (
      <div style={{ color: "red", padding: "2rem", textAlign: "center" }}>
        <h1>An error occurred</h1>
        <p>{error.message}</p> 
      </div>
    )
  }
}

function getErrorBoundaryHandlers(
    props?: BaseErrorBoundaryProps
  ): Required<BaseErrorBoundaryProps> {
    return {
      routeErrorConfig: {
        fn: props?.routeErrorConfig?.fn ?? defaultErrorBoundaryProps.routeErrorConfig.fn,
        render: props?.routeErrorConfig?.render ?? defaultErrorBoundaryProps.routeErrorConfig.render,
      },
      unexpectedErrorConfig: {
        fn: props?.unexpectedErrorConfig?.fn ?? defaultErrorBoundaryProps.unexpectedErrorConfig.fn,
        render: props?.unexpectedErrorConfig?.render ?? defaultErrorBoundaryProps.unexpectedErrorConfig.render,
      },
    };
  }

export function BaseErrorBoundary(props?: BaseErrorBoundaryProps): React.ReactElement {
  const { routeErrorConfig, unexpectedErrorConfig } = getErrorBoundaryHandlers(props);
  const error = useRouteError();

  if (isRouteErrorResponse(error)) {
    if (routeErrorConfig.fn) {
      routeErrorConfig.fn(error);
    }

    return routeErrorConfig.render(error);
  }

  if (unexpectedErrorConfig.fn) {
    unexpectedErrorConfig.fn(error as Error);
  }

  return unexpectedErrorConfig.render(error as Error);
}