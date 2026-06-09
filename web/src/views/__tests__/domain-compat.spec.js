import { describe, expect, it, vi } from 'vitest';

import AuthEditSetting from '../../components/auth/EditSetting.vue';
import AuthLabelItem from '../../components/auth/LabelItem.vue';
import PolicyEditSetting from '../../components/policy/EditSetting.vue';
import PolicyLabelItem from '../../components/policy/LabelItem.vue';
import Requestauth from '../auth/Requestauth.vue';
import UserView from '../user/User.vue';

describe('domain Vue 3 migration helpers', () => {
  it('defines native request authentication list columns', () => {
    const columns = Requestauth.computed.columns.call({
      $t: (key) => key
    });

    expect(columns.map((column) => column.key)).toEqual([
      'name',
      'labels',
      'namespace',
      'ruleCount',
      'createdAt'
    ]);
    expect(columns[0].link).toBe(true);
  });

  it('resets auth request status when the auth request editor mounts', async () => {
    const commit = vi.fn();

    await AuthEditSetting.mounted.call({
      $store: {
        commit,
        dispatch: vi.fn(async () => null)
      },
      $route: {
        query: {
          name: 'auth',
          namespace: 'default'
        }
      },
      fetchData: vi.fn(async () => null)
    });

    expect(commit).toHaveBeenCalledWith('AuthRequest_ResetStatus');
  });

  it('removes auth labels without sending undefined payload fields', () => {
    const commit = vi.fn();
    const vm = {
      index: 2,
      labels: [{}, {}, {}],
      $store: {
        commit
      }
    };

    AuthLabelItem.methods.removeLabel.call(vm);

    expect(commit).toHaveBeenCalledWith('AuthRequest_RemoveLabel', {
      index: 2
    });
  });

  it('removes policy labels without sending undefined payload fields', () => {
    const commit = vi.fn();
    const vm = {
      index: 1,
      $store: {
        commit
      }
    };

    PolicyLabelItem.methods.removeLabel.call(vm);

    expect(commit).toHaveBeenCalledWith('AuthPolicy_RemoveLabel', {
      index: 1
    });
  });

  it('uses the valid translation key for invalid policy names', () => {
    const errors = PolicyEditSetting.computed.nameErrors.call({
      submitted: true,
      touched: {
        name: true
      },
      name: 'invalid name!',
      $t: (key) => key
    });

    expect(errors).toContain('Form.Valid');
  });

  it('builds user detail links from stable user IDs', () => {
    expect(UserView.methods.detailUrl({ id: 'user-1' })).toBe('/user/user-1');
    expect(UserView.methods.rowKey({ id: 'user-1', username: 'admin' })).toBe('user-1');
  });
});
