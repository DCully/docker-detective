import axios from 'axios';

const BASE_API_URI = ''

export const getName = async () => {
    const resp = await axios.get(`${BASE_API_URI}/name`)
    return resp.data.imageName
}

export const getFileSystems = async () => {
    const resp = await axios.get(`${BASE_API_URI}/filesystems`)
    return resp.data
}

export const getDirData = async(dirId: number) => {
    const resp = await axios.get(`${BASE_API_URI}/dirData?id=${dirId}`)
    return resp.data
}
