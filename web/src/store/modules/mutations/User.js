const User_ResetStatus = (state) => {
  state.status = '';
  state.error_handle = '';
}

const User_SetStatus = (state, payload) => {
  state.status = payload.status;
  state.error_handle = payload.error_handle;
}

const User_GetItem = (state, payload) => {
  state.user = payload.user;
}

const User_GetItems = (state, payload) => {
  let users = [];
  if (payload.users) {
    for (let index in payload.users) {
      users.push({
        id: payload.users[index].uid,
        name: payload.users[index].name,
        username: payload.users[index].username,
        permissions: payload.users[index].permissions,
        email: payload.users[index].email,
        isDisabled: payload.users[index].isDisabled,
        createdAt: payload.users[index].createdAt,
      })
    }
  }

  state.users = users;
  state.meta = payload.meta;
}

export default {
  User_ResetStatus,
  User_SetStatus,
  User_GetItem,
  User_GetItems,
}
