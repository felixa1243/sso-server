export const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

async function request(method: string, url: string, token: string | null = null, body: any = null) {
    const headers: HeadersInit = {
        'Content-Type': 'application/json',
    };

    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }

    const options: RequestInit = {
        method,
        headers,
    };

    if (body) {
        options.body = JSON.stringify(body);
    }

    try {
        const response = await fetch(`${API_BASE_URL}${url}`, options);
        const text = await response.text();

        let data;
        try {
            data = JSON.parse(text);
        } catch (e) {
            // If response is not JSON, it might be an HTML error page or empty
            if (!response.ok) {
                throw new Error(text || `Request failed with status ${response.status}`);
            }
            return text; // Return text if successful but not JSON
        }

        if (!response.ok) {
            throw new Error(data.message || data.error || `Request failed with status ${response.status}`);
        }

        return data;
    } catch (error: any) {
        console.error('API Error:', error);
        throw error;
    }
}

export const get = (url: string, token: string | null) => request('GET', url, token);
export const post = (url: string, body: any, token: string | null) => request('POST', url, token, body);
export const put = (url: string, token: string | null, body: any = null) => request('PUT', url, token, body);
export const del = (url: string, token: string | null) => request('DELETE', url, token);
