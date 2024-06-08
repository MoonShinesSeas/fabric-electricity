// import { login, logout, getInfo } from '@/api/user'
// import { getToken, setToken, removeToken } from '@/utils/auth'
// import { resetRouter } from '@/router'

// const getDefaultState = () => {
//   return {
//     token: getToken(),
//     name: '',
//     id: '',
//     introduction: '',
//     avatar: ''
//   }
// }

// const state = getDefaultState()

// const mutations = {
//   RESET_STATE: (state) => {
//     Object.assign(state, getDefaultState())
//   },
//   SET_TOKEN: (state, token) => {
//     state.token = token
//   },
//   SET_NAME: (state, name) => {
//     state.name = name
//   },
//   SET_ID: (state, id) => {
//     state.id = id
//   },
//   SET_INTRODUCTION: (state, introduction) => {
//     state.introduction = introduction
//   },
//   SET_AVATAR: (state, avatar) => {
//     state.avatar = avatar
//   }
// }

// const actions = {
//   // user login
//   login({ commit }, userInfo) {
//     const { username, password } = userInfo
//     return new Promise((resolve, reject) => {
//       login({ username: username.trim(), password: password }).then(response => {
//         const { data } = response
//         commit('SET_TOKEN', data.token)
//         setToken(data.token)
//         resolve()
//       }).catch(error => {
//         reject(error)
//       })
//     })
//   },

//   // get user info
//   getInfo({ commit, state }) {
//     return new Promise((resolve, reject) => {
//       getInfo(state.token).then(response => {
//         const { data } = response

//         if (!data) {
//           reject('Verification failed, please Login again.')
//         }

//         const { name, id, avatar, introduction } = data

//         commit('SET_NAME', name)
//         commit('SET_ID', id)
//         commit('SET_INTRODUCTION', introduction)
//         commit('SET_AVATAR', avatar)
//         resolve(data)
//       }).catch(error => {
//         reject(error)
//       })
//     })
//   },

//   // user logout
//   logout({ commit, state }) {
//     return new Promise((resolve, reject) => {
//       logout(state.token).then(() => {
//         removeToken() // must remove  token  first
//         resetRouter()
//         commit('RESET_STATE')
//         resolve()
//       }).catch(error => {
//         reject(error)
//       })
//     })
//   },

//   // remove token
//   resetToken({ commit }) {
//     return new Promise(resolve => {
//       removeToken() // must remove  token  first
//       commit('RESET_STATE')
//       resolve()
//     })
//   }
// }

// export default {
//   namespaced: true,
//   state,
//   mutations,
//   actions
// }

// import { getwallet } from '@/api/user'  
  
// const state = {  
//   username: 'Alice',  
//   address: '', 
//   balance: 0  
// }  
  
// const mutations = {  
//   SET_USERNAME: (state, username) => {  
//     state.username = username  
//   },  
//   SET_ADDRESS: (state, address) => {  
//     state.address = address  
//   },  
//   SET_BALANCE: (state, balance) => {  
//     state.balance = balance  
//   }  
// }  
  
// const actions = {  
//   // 获取用户信息  
//   getwallet({ commit }) {  
//     return new Promise((resolve, reject) => {  
//       getwallet().then(response => {  
//         const { data } = response  
          
//         if (!data) {  
//           reject('Failed to fetch user info.')  
//         }  
  
//         const { address, balance } = data // 假设后端返回的数据结构包含这些字段  
  
//         commit('SET_USERNAME', username)  
//         commit('SET_ADDRESS', address)  
//         commit('SET_BALANCE', balance)  
//         resolve(data)  
//       }).catch(error => {  
//         reject(error)  
//       })  
//     })  
//   }  
// }  
  
// export default {  
//   namespaced: true,  
//   state,  
//   mutations,  
//   actions  
// }