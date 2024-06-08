import request from '@/utils/request'

export function listOrder(data){
  return request({
    url: '/proposal/getProposalBySeller',
    method: 'post',
    data
  })
}

export function getProposalByBuyer(data){
  return request({
    url: '/proposal/getProposalByBuyer',
    method: 'post',
    data
  })
}


export function submitOrder(data){
  return request({
    url: '/user/submitOrder',
    method: 'post',
    data
  })
}

export function addOrder(data){
  return request({
    url: '/proposal/setProposal',
    method: 'post',
    data
  })
}


export function updateOrder(data){
  return request({
    url: '/proposal/updateProposal',
    method: 'post',
    data
  })
}


export function deleteOrder(id){
  return request({
    url: '/order/delete/' + id,
    method: 'post'
  })
}

export function searchOrderByBuyer(data){
  return request({
    url: '/proposal/getProposalBySeller',
    method: 'post',
    data
  })
}
