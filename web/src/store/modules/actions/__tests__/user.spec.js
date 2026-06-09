import { beforeEach, describe, expect, it, vi } from 'vitest';

const axiosMock = {
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  delete: vi.fn(),
  defaults: {
    headers: {
      common: {},
      post: {}
    }
  }
};

vi.mock('axios', () => ({
  default: axiosMock
}));

describe('User actions', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    global.sessionStorage = {
      getItem: vi.fn(() => 'token')
    };
  });

  it('commits the standard error status mutation when user creation fails', async () => {
    axiosMock.post.mockRejectedValueOnce({
      response: {
        data: 'create failed'
      }
    });

    const { default: userActions } = await import('../User.js');
    const commit = vi.fn();

    await userActions.User_NewItem({ commit }, {
      name: 'Pilotwave Admin',
      username: 'admin',
      password: 'secret',
      email: 'admin@example.com',
      permissions: ['admin']
    });

    expect(commit).toHaveBeenCalledWith('User_SetStatus', {
      status: 'create_error',
      error_handle: 'create failed'
    });
  });
});
