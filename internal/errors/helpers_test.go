// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0
package errors

import (
	"errors"
	"net/http"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHelpers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Helpers Suite")
}

var _ = Describe("Error Helpers", func() {
	Describe("IsNotFound", func() {
		Context("when error is NotFound", func() {
			It("should return true", func() {
				err := E(Op("test"), NotFound)
				Expect(IsNotFound(err)).To(BeTrue())
			})
		})

		Context("when error is not NotFound", func() {
			It("should return false for InvalidArgument", func() {
				err := E(Op("test"), InvalidArgument)
				Expect(IsNotFound(err)).To(BeFalse())
			})

			It("should return false for non-Error type", func() {
				err := errors.New("standard error")
				Expect(IsNotFound(err)).To(BeFalse())
			})

			It("should return false for nil error", func() {
				Expect(IsNotFound(nil)).To(BeFalse())
			})
		})
	})

	Describe("IsAlreadyExists", func() {
		Context("when error is AlreadyExists", func() {
			It("should return true", func() {
				err := E(Op("test"), AlreadyExists)
				Expect(IsAlreadyExists(err)).To(BeTrue())
			})
		})

		Context("when error is not AlreadyExists", func() {
			It("should return false for NotFound", func() {
				err := E(Op("test"), NotFound)
				Expect(IsAlreadyExists(err)).To(BeFalse())
			})

			It("should return false for non-Error type", func() {
				err := errors.New("standard error")
				Expect(IsAlreadyExists(err)).To(BeFalse())
			})

			It("should return false for nil error", func() {
				Expect(IsAlreadyExists(nil)).To(BeFalse())
			})
		})
	})

	Describe("IsInvalidArgument", func() {
		Context("when error is InvalidArgument", func() {
			It("should return true", func() {
				err := E(Op("test"), InvalidArgument)
				Expect(IsInvalidArgument(err)).To(BeTrue())
			})
		})

		Context("when error is not InvalidArgument", func() {
			It("should return false for NotFound", func() {
				err := E(Op("test"), NotFound)
				Expect(IsInvalidArgument(err)).To(BeFalse())
			})

			It("should return false for non-Error type", func() {
				err := errors.New("standard error")
				Expect(IsInvalidArgument(err)).To(BeFalse())
			})

			It("should return false for nil error", func() {
				Expect(IsInvalidArgument(nil)).To(BeFalse())
			})
		})
	})

	Describe("IsInternal", func() {
		Context("when error is Internal", func() {
			It("should return true", func() {
				err := E(Op("test"), Internal)
				Expect(Is(err, Internal)).To(BeTrue())
			})
		})

		Context("when error is not Internal", func() {
			It("should return false for NotFound", func() {
				err := E(Op("test"), NotFound)
				Expect(Is(err, Internal)).To(BeFalse())
			})

			It("should return false for non-Error type", func() {
				err := errors.New("standard error")
				Expect(Is(err, Internal)).To(BeFalse())
			})

			It("should return false for nil error", func() {
				Expect(Is(nil, Internal)).To(BeFalse())
			})
		})
	})

	Describe("CodeToHTTPStatus", func() {
		Context("with standard error codes", func() {
			It("should map OK to 200", func() {
				Expect(CodeToHTTPStatus(OK)).To(Equal(http.StatusOK))
			})

			It("should map InvalidArgument to 400", func() {
				Expect(CodeToHTTPStatus(InvalidArgument)).To(Equal(http.StatusBadRequest))
			})

			It("should map Unauthenticated to 401", func() {
				Expect(CodeToHTTPStatus(Unauthenticated)).To(Equal(http.StatusUnauthorized))
			})

			It("should map PermissionDenied to 403", func() {
				Expect(CodeToHTTPStatus(PermissionDenied)).To(Equal(http.StatusForbidden))
			})

			It("should map NotFound to 404", func() {
				Expect(CodeToHTTPStatus(NotFound)).To(Equal(http.StatusNotFound))
			})

			It("should map AlreadyExists to 409", func() {
				Expect(CodeToHTTPStatus(AlreadyExists)).To(Equal(http.StatusConflict))
			})

			It("should map Internal to 500", func() {
				Expect(CodeToHTTPStatus(Internal)).To(Equal(http.StatusInternalServerError))
			})
		})

		Context("with unknown error code", func() {
			It("should default to 500", func() {
				unknownCode := Code("UNKNOWN_CODE")
				Expect(CodeToHTTPStatus(unknownCode)).To(Equal(http.StatusInternalServerError))
			})
		})
	})

	Describe("Error constructors", func() {
		Describe("NotFoundError", func() {
			It("should create NotFound error with entity and ID", func() {
				err := NotFoundError("test.op", "User", "123")

				Expect(err).NotTo(BeNil())
				Expect(IsNotFound(err)).To(BeTrue())

				appErr := err
				Expect(appErr.Code).To(Equal(NotFound))
				Expect(appErr.Op).To(Equal("test.op"))
				Expect(appErr.Entity).To(Equal("User"))
				Expect(appErr.ID).To(Equal("123"))
			})
		})

		Describe("AlreadyExistsError", func() {
			It("should create AlreadyExists error with entity and ID", func() {
				err := AlreadyExistsError("test.op", "User", "123")

				Expect(err).NotTo(BeNil())
				Expect(IsAlreadyExists(err)).To(BeTrue())

				appErr := err
				Expect(appErr.Code).To(Equal(AlreadyExists))
				Expect(appErr.Op).To(Equal("test.op"))
				Expect(appErr.Entity).To(Equal("User"))
				Expect(appErr.ID).To(Equal("123"))
			})
		})

		Describe("InvalidArgumentError", func() {
			It("should create InvalidArgument error with message", func() {
				err := InvalidArgumentError("test.op", "User", "invalid input")

				Expect(err).NotTo(BeNil())
				Expect(IsInvalidArgument(err)).To(BeTrue())

				appErr := err
				Expect(appErr.Code).To(Equal(InvalidArgument))
				Expect(appErr.Op).To(Equal("test.op"))
				Expect(appErr.Entity).To(Equal("User"))
				Expect(appErr.Message).To(Equal("invalid input"))
			})
		})

		Describe("InternalError", func() {
			It("should create Internal error wrapping underlying error", func() {
				underlyingErr := errors.New("database connection failed")
				err := InternalError("test.op", "User", "123", underlyingErr)

				Expect(err).NotTo(BeNil())
				Expect(Is(err, Internal)).To(BeTrue())

				appErr := err
				Expect(appErr.Code).To(Equal(Internal))
				Expect(appErr.Op).To(Equal("test.op"))
				Expect(appErr.Entity).To(Equal("User"))
				Expect(appErr.ID).To(Equal("123"))
				Expect(appErr.Err).To(Equal(underlyingErr))
			})
		})
	})

	Describe("Integration tests", func() {
		Context("when chaining error operations", func() {
			It("should preserve error types through helper functions", func() {
				// Create a NotFound error
				err := NotFoundError("service.Get", "User", "456")

				// Verify it's detected correctly
				Expect(IsNotFound(err)).To(BeTrue())
				Expect(IsAlreadyExists(err)).To(BeFalse())
				Expect(IsInvalidArgument(err)).To(BeFalse())
				Expect(Is(err, Internal)).To(BeFalse())

				// Verify HTTP mapping
				Expect(CodeToHTTPStatus(NotFound)).To(Equal(http.StatusNotFound))
			})
		})

		Context("when working with error hierarchy", func() {
			It("should handle wrapped errors correctly", func() {
				originalErr := errors.New("connection timeout")
				wrappedErr := InternalError("db.Query", "Connection", "", originalErr)

				// Should detect as Internal error
				Expect(Is(wrappedErr, Internal)).To(BeTrue())

				// Should preserve original error
				appErr := wrappedErr
				Expect(appErr.Err).To(Equal(originalErr))
			})
		})
	})
})
