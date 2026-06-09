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

const flushPromises = () => new Promise((resolve) => setTimeout(resolve, 0));

describe('User action API contract', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.resetModules();
    global.sessionStorage = {
      getItem: vi.fn(() => 'token')
    };
  });

  it('sends create user payload', async () => {
    axiosMock.post.mockResolvedValueOnce({ status: 200, data: {} });

    const { default: userActions } = await import('../User.js');
    const commit = vi.fn();

    userActions.User_NewItem({ commit }, {
      name: 'Pilotwave Admin',
      username: 'admin',
      password: 'admin',
      email: 'admin@example.com',
      permissions: ['admin']
    });
    await flushPromises();

    expect(axiosMock.post).toHaveBeenCalledWith(
      expect.stringContaining('/users'),
      {
        name: 'Pilotwave Admin',
        username: 'admin',
        password: 'admin',
        email: 'admin@example.com',
        permissions: ['admin']
      }
    );
  });

  it('sends update user payload without password', async () => {
    axiosMock.put.mockResolvedValueOnce({ status: 200, data: {} });

    const { default: userActions } = await import('../User.js');
    const commit = vi.fn();

    userActions.User_UpdateItem({ commit }, {
      id: 'user-1',
      name: 'Pilotwave Admin',
      email: 'admin@example.com',
      permissions: ['admin']
    });
    await flushPromises();

    expect(axiosMock.put).toHaveBeenCalledWith(
      expect.stringContaining('/user/user-1'),
      {
        name: 'Pilotwave Admin',
        email: 'admin@example.com',
        permissions: ['admin']
      }
    );
  });

  it('sends reset password payload', async () => {
    axiosMock.put.mockResolvedValueOnce({ status: 200, data: {} });

    const { default: userActions } = await import('../User.js');
    const commit = vi.fn();

    userActions.User_UpdatePwd({ commit }, {
      id: 'user-1',
      password: 'admin'
    });
    await flushPromises();

    expect(axiosMock.put).toHaveBeenCalledWith(
      expect.stringContaining('/user/user-1/resetpassword'),
      {
        password: 'admin'
      }
    );
  });

  it('deletes users by id', async () => {
    axiosMock.delete.mockResolvedValueOnce({ status: 200, data: {} });

    const { default: userActions } = await import('../User.js');
    const commit = vi.fn();

    userActions.User_DelItem({ commit }, {
      id: 'user-1'
    });
    await flushPromises();

    expect(axiosMock.delete).toHaveBeenCalledWith(expect.stringContaining('/user/user-1'));
  });
});
