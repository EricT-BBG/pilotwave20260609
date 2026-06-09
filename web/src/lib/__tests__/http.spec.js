import { beforeEach, describe, expect, it, vi } from 'vitest';

const responseUse = vi.fn();
const axiosMock = {
  create: vi.fn(() => ({ interceptors: { request: { use: vi.fn() } } })),
  defaults: {},
  interceptors: {
    response: {
      use: responseUse,
    },
  },
};

vi.mock('axios', () => ({
  default: axiosMock,
}));

describe('http client defaults', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.resetModules();
    axiosMock.defaults = {};
  });

  it('sets a finite timeout for API calls', async () => {
    await import('../http');

    expect(axiosMock.defaults.timeout).toBe(15000);
    expect(axiosMock.create).toHaveBeenCalledWith(expect.objectContaining({
      timeout: 15000,
    }));
  });

  it('emits a visible API error event when the backend does not respond', async () => {
    const dispatchEvent = vi.fn();
    vi.stubGlobal('window', {
      dispatchEvent,
    });
    vi.stubGlobal('CustomEvent', class {
      constructor(type, options) {
        this.type = type;
        this.detail = options.detail;
      }
    });

    await import('../http');
    const [, onRejected] = responseUse.mock.calls[0];
    const error = new Error('timeout of 15000ms exceeded');
    error.code = 'ECONNABORTED';

    await expect(onRejected(error)).rejects.toBe(error);

    expect(dispatchEvent).toHaveBeenCalledWith(expect.objectContaining({
      type: 'pilotwave-api-error',
      detail: expect.objectContaining({
        message: 'Alert.ApiUnavailable',
      }),
    }));

    vi.unstubAllGlobals();
  });
});
