import { describe, expect, it } from 'vitest';
import { splitTLSPEM, summarizeTLSPEM } from '../pem.js';

const cert = `-----BEGIN CERTIFICATE-----
MIIB
-----END CERTIFICATE-----
`;

const key = `-----BEGIN RSA PRIVATE KEY-----
MIIE
-----END RSA PRIVATE KEY-----
`;

describe('PEM helpers', () => {
  it('splits combined certificate and private key PEM', () => {
    expect(splitTLSPEM(cert + key)).toMatchObject({
      cert,
      pkey: key,
      certificateCount: 1,
      privateKeyCount: 1,
    });
  });

  it('summarizes TLS material without exposing raw key data', () => {
    expect(summarizeTLSPEM({ cert, pkey: key, cacert: cert })).toEqual({
      certificateCount: 1,
      hasPrivateKey: true,
      caCertificateCount: 1,
    });
  });

  it('keeps certificate and private key recognition tied to their own fields', () => {
    expect(summarizeTLSPEM({ cert: '', pkey: cert + key })).toEqual({
      certificateCount: 0,
      hasPrivateKey: true,
      caCertificateCount: 0,
    });

    expect(summarizeTLSPEM({ cert: cert + key, pkey: '' })).toEqual({
      certificateCount: 1,
      hasPrivateKey: false,
      caCertificateCount: 0,
    });
  });
});
