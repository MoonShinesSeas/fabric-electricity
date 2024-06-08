import request from '@/utils/request'

export function listElec(){
  return request({
    url: '/good/getAll',
    method: 'get'
  })
}
export function addElec(data){
  return request({
    url: '/elec/add',
    method: 'post',
    data
  })
}

export function updateElec(data){
  return request({
    url: '/good/updateGoodPrice',
    method: 'post',
    data
  })
}

export function deleteElec(id){
  return request({
    url: '/elec/delete/' + id,
    method: 'post'
  })
}

export function searchElec(data){
  return request({
    url: '/good/getGoodByOwner',
    method: 'post',
    data
  })
}
