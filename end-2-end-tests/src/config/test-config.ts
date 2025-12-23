/**
 * E2E Test Configuration
 *
 * Configuration for E2E tests targeting the actual microservice APIs.
 * Values are loaded from environment variables with sensible defaults.
 */

export interface ServiceConfig {
  url: string;
  username?: string;
  password?: string;
}

export interface TestConfig {
  baseUrl: string;
  services: {
    billing: ServiceConfig;
    kitchen: ServiceConfig;
    reservations: ServiceConfig;
    roomservice: ServiceConfig;
    lobby: ServiceConfig;
    cleaning: ServiceConfig;
  };
  timeouts: {
    default: number;
  };
}

const getEnvOrDefault = (key: string, defaultValue: string): string => {
  return process.env[key] || defaultValue;
};

const getBaseUrl = (): string => {
  return getEnvOrDefault("STAGING_BASE_URL", "https://chotel.cuini.me");
};

/**
 * Get service URL for E2E testing.
 * Uses external ingress URL for all environments.
 */
const getServiceUrl = (serviceName: string): string => {
  return `${getBaseUrl()}/api/${serviceName}`;
};

export const config: TestConfig = {
  baseUrl: getBaseUrl(),
  services: {
    billing: {
      url: getServiceUrl("billing"),
    },
    kitchen: {
      url: getServiceUrl("kitchen"),
    },
    reservations: {
      url: getServiceUrl("reservations"),
    },
    roomservice: {
      url: getServiceUrl("roomservice"),
    },
    lobby: {
      url: getServiceUrl("lobby"),
    },
    cleaning: {
      url: getServiceUrl("cleaning"),
      username: getEnvOrDefault("CLEANING_SERVICE_USER", "Marcelo"),
      password: getEnvOrDefault("CLEANING_SERVICE_PASSWORD", "chotardo69"),
    },
  },
  timeouts: {
    default: parseInt(getEnvOrDefault("TEST_TIMEOUT_DEFAULT", "15000")),
  },
};

export default config;
